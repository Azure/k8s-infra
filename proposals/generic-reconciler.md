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
  * Process - need to give background on motivation (lack of etags, lack of consistent ATCH etc)
    * Store `last-applied` state in CRD (annotation?)
    * when reconciling, diff `spec` and `last-applied` to generate a patch metadata
    * if ARM API supports `PATCH` then use then generate a patch with the metadata
    * if ARM API doesn't support patch then `GET` and apply patch metadata to the response, hen use this to issue a `PUT`
    * Apply changes in patch metadata to `last-applied`
    * Issue a `GET` and use this to update latest spec/status
    * Add an example here for VMSS (pseudo-CRD with flow of applying updates)
