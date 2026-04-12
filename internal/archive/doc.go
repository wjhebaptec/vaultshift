// Package archive provides versioned secret archiving for vaultshift.
//
// The Archive type stores historical values for secrets across any provider,
// retaining up to a configurable number of recent entries per key. This
// enables audit trails, rollback support, and drift analysis over time.
//
// Basic usage:
//
//	 a := archive.New(10) // keep last 10 versions per key
//	 a.Store("aws", "db/password", "s3cr3t")
//	 a.Store("aws", "db/password", "n3ws3cr3t")
//
//	 latest, err := a.Latest("aws", "db/password")
//	 history := a.List("aws", "db/password")
//	 a.Purge("aws", "db/password")
package archive
