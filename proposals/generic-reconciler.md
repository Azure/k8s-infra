# Generic reconciler

TODO

 * Diff CREATE/PUT/GET to build up create-only, immutable and writeable field lists
 * Use swagger to generate structs
 * Generic validator driven off generated structs
   * enforces create-only and immutable fields not being changed
   * allow pluggable validation hooks for custom logic if needed
 * Generic reconciler driven off generated structs
   * generic loop for CRD reconciliation
   * allow pluggable mappers/mutators as needed



