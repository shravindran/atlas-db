package main

import (
	"fmt"

	"strings"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	atlas "github.com/infobloxopen/atlas-db/pkg/apis/db/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

var dbDriverMap = make(map[string]database.Driver)

const (
	badConn    = "bad connection"
	connClosed = "connection is already closed"
)

func (c *Controller) enqueueDatabaseSchema(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		c.logger.Debug("not enqueue schema object")
		return
	}
	c.logger.Infof("enqueue schema object: %s", object.GetName())
	c.enqueue(obj, c.schemaQueue)
}

func (c *Controller) syncSchema(key string) error {
	c.logger.Infof("Processing schema : %v", key)

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		c.logger.Errorf("invalid resource key: %s", key)
		return err
	}

	// Get the Schema resource with this namespace/name
	schema, err := c.schemasLister.DatabaseSchemas(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			c.logger.Warningf("schema '%s' in work queue no longer exists", key)
			return nil
		}
		return err
	}
	c.logger.Debugf("Schema Spec: %v", schema.Spec)

	// If dsn/dsnFrom is passed in the schema spec consider as override and don't go through database spec
	dsn := schema.Spec.Dsn
	dbName := schema.Spec.Database
	if dsn == "" {
		if schema.Spec.DsnFrom != nil {
			secretName := schema.Spec.DsnFrom.SecretKeyRef.Name
			dsn, err = c.getSecretFromValueSource(schema.Namespace, schema.Spec.DsnFrom)
			if err != nil {
				if errors.IsNotFound(err) {
					msg := fmt.Sprintf(MessageDSNGetWaiting, key, secretName)
					c.logger.Debug(msg)
					c.updateDatabaseSchemaStatus(key, schema, StatePending, msg)
					return err
				}
				msg := fmt.Sprintf(MessageDSNGetFailure, key, secretName, err)
				c.logger.Error(schemaStatusMsg)
				c.updateDatabaseSchemaStatus(key, schema, StateError, msg)
				return nil
			}
		} else { // Get the dsn from database created secret
			db, err := c.dbsLister.Databases(namespace).Get(dbName)
			if err != nil {
				schemaStatusMsg = fmt.Sprintf("failed to fetch database info `%s`: %s", dbName, err)
				c.logger.Error(schemaStatusMsg)
				c.updateDatabaseSchemaStatus(key, schema, StateError, schemaStatusMsg)
				return err
			}

			dsn, err = c.getSecretByName(db.Namespace, "dsn", dbName)
			if err != nil {
				if errors.IsNotFound(err) {
					msg := fmt.Sprintf(MessageDSNGetWaiting, key, dbName)
					c.logger.Debug(msg)
					c.updateDatabaseSchemaStatus(key, schema, StatePending, msg)
					return err
				}
				msg := fmt.Sprintf(MessageDSNGetFailure, key, dbName, err)
				c.logger.Warningf(msg)
				c.updateDatabaseSchemaStatus(key, schema, StateError, msg)
				return nil
			}
		}
	}

	// Formulate sourceURL from either source string or secret provided
	sourceURL := schema.Spec.Source
	if sourceURL == "" {
		if schema.Spec.SourceFrom != nil {
			secretName := schema.Spec.SourceFrom.SecretKeyRef.Name
			sourceURL, err = c.getSecretFromValueSource(schema.Namespace, schema.Spec.SourceFrom)
			if err != nil {
				if errors.IsNotFound(err) {
					msg := fmt.Sprintf("waiting to get sourceURL for schema `%s` from secret `%s`", key, secretName)
					c.logger.Debug(schemaStatusMsg)
					c.updateDatabaseSchemaStatus(key, schema, StatePending, msg)
					return err
				}
				msg := fmt.Sprintf("failed to get valid sourceURL for schema `%s` from secret `%s`", key, secretName)
				c.logger.Warningf(msg)
				c.updateDatabaseSchemaStatus(key, schema, StateError, msg)
				return nil
			}

		} else {
			msg := fmt.Sprintf("failed to get valid sourceURL for schema `%s`", key)
			c.logger.Error(msg)
			c.updateDatabaseSchemaStatus(key, schema, StateError, msg)
			return err
		}
	}

	// migrate package is not closing the dbconnnection so using a local cache reuse dbconnection.
	// TODO when same resource with different dsn arrives; need to unset local cache during schema resource deletion.
	dbKey := schema.Namespace + "." + schema.Name
	dbDriver, ok := dbDriverMap[dbKey]
	if !ok {
		dbDriver, err = database.Open(dsn)
		if err != nil {
			schemaStatusMsg = fmt.Sprintf("failed to open dbconnection: %s", err)
			c.logger.Error(schemaStatusMsg)
			c.updateDatabaseSchemaStatus(key, schema, StateError, schemaStatusMsg)
			return err
		}
		dbDriverMap[dbKey] = dbDriver
	}

	mgrt, err := migrate.NewWithDatabaseInstance(sourceURL, dbName, dbDriver)
	if err != nil {
		schemaStatusMsg = fmt.Sprintf("failed to initialize migrate engine: %s", err)
		c.logger.Error(schemaStatusMsg)
		c.updateDatabaseSchemaStatus(key, schema, StateError, schemaStatusMsg)
		return err
	}

	ver, dirt, err := mgrt.Version()
	if err != nil {
		errString := err.Error()
		if strings.Contains(errString, badConn) || strings.Contains(errString, connClosed) {
			delete(dbDriverMap, dbKey)
			schemaStatusMsg = fmt.Sprintf("database connection errorred: %s", err)
			c.logger.Error(schemaStatusMsg)
			c.updateDatabaseSchemaStatus(key, schema, StateError, schemaStatusMsg)
			return err
		} else if err == migrate.ErrNilVersion {
			schemaStatusMsg = fmt.Sprintf("database `%s` has no migration applied", dbName)
			c.logger.Infof(schemaStatusMsg)
			c.updateDatabaseSchemaStatus(key, schema, StatePending, schemaStatusMsg)
		} else {
			schemaStatusMsg = fmt.Sprintf("cannot get current database version: %s", err)
			c.logger.Error(schemaStatusMsg)
			c.updateDatabaseSchemaStatus(key, schema, StateError, schemaStatusMsg)
			return err
		}
	}
	if dirt {
		// TODO we might want to notficate someone about this
		schemaStatusMsg = fmt.Sprintf("database `%s` is in dirty state (current version is %d)", dbName, ver)
		c.logger.Error(schemaStatusMsg)
		c.updateDatabaseSchemaStatus(key, schema, StateError, schemaStatusMsg)
		return err
	}

	toVersion := uint(schema.Spec.Version)
	if ver == toVersion {
		schemaStatusMsg = fmt.Sprintf("Databaseschema `%s` is in requested version %d", key, toVersion)
		c.logger.Infof(schemaStatusMsg)
		c.updateDatabaseSchemaStatus(key, schema, StateSuccess, fmt.Sprintf(MessageSchemaSynced, key))
		return nil
	}

	err = mgrt.Migrate(toVersion)
	if err != nil {
		schemaStatusMsg = fmt.Sprintf("cannot migrate the db %s : %s", dbName, err)
		c.logger.Error(schemaStatusMsg)
		c.updateDatabaseSchemaStatus(key, schema, StateError, schemaStatusMsg)
		return err
	}
	msg := fmt.Sprintf("Migration successful from version %d to %d", ver, toVersion)
	c.recorder.Event(schema, corev1.EventTypeNormal, SuccessSynced, msg)

	c.updateDatabaseSchemaStatus(key, schema, StateSuccess, fmt.Sprintf(MessageSchemaSynced, key))
	return nil
}

func (c *Controller) updateDatabaseSchemaStatus(key string, schema *atlas.DatabaseSchema, state, msg string) (*atlas.DatabaseSchema, error) {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	schemaCopy := schema.DeepCopy()
	schemaCopy.Status.State = state
	schemaCopy.Status.Message = msg
	// Until #38113 is merged, we must use Update instead of UpdateStatus to
	// update the Status block of the resource. UpdateStatus will not
	// allow changes to the Spec of the resource, which is ideal for ensuring
	// nothing other than resource status has been updated.

	_, err := c.atlasclientset.AtlasdbV1alpha1().DatabaseSchemas(schema.Namespace).Update(schemaCopy)
	if err != nil {
		c.logger.Warningf("error updating status to '%s' for database schema '%s': %s", state, key, err)
		return schema, err
	}
	// we have to pull it back out or our next update will fail. hopefully this is fixed with updateStatus
	return c.schemasLister.DatabaseSchemas(schema.Namespace).Get(schema.Name)
}
