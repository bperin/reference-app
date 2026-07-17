# Content Storage Architecture

## Content Storage Boundary
The content storage domain manages the lifecycle of user-uploaded files, ensuring integrity and security through GCS and the database.

## Signed URL Workflow
1. Client requests an upload URL from the API.
2. The service validates user permissions.
3. The service generates a signed GCS URL with restricted access (limited time, specific object path).
4. The client uploads the file directly to GCS.

## Machine-Identity Requirements for Eventarc
Eventarc triggers require a service account with the necessary permissions:
- `roles/run.invoker`: For triggering the Cloud Run callback endpoint.
- Storage IAM roles: For managing files in GCS if required by the service account's role.

## Idempotency Rules for Completion Callbacks
Callbacks MUST be idempotent. The API check the object ID against the database for an existing processing state before acting.

## Error-Handling Policy
- **Stale generation**: Files that remain in GCS without a corresponding finalized record after a threshold (e.g., 24 hours) are considered stale and should be cleaned up.
- **Unknown objects**: Objects in GCS without a corresponding record or in an unexpected state should be ignored and logged for audit; do not attempt to process them.
