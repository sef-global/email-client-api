Send Bulk Email API using GO 

![Email API](https://github.com/mayura-andrew/send-bulk-email-client-api/assets/48531182/243b444d-9284-4008-be17-f678fe602c75)


# Email Service API

This API provides functionality for sending, tracking, and querying emails.

## Endpoints

### POST /sendemail (Status : Completed ☑️)

Sends an email. The request body should be a JSON object with the following fields:

- `recipient`: The email address to send the email to.
- `subject`: The subject of the email.
- `body`: The body of the email.

### POST /trackemail  (Status : Inprogress ⏳)

Updates the tracking status of an email. The request body should be a JSON object with the following fields:

- `id`: The ID of the email to track.
- `opened`: A boolean indicating whether the email has been opened.

### GET /successfullysent (Status : Not Started ❌)

Retrieves all emails that have been successfully sent. Returns a JSON array of email objects.

### GET /notsuccessfullysent  (Status : Not Started ❌)

Retrieves all emails that have not been successfully sent. Returns a JSON array of email objects.

### GET /totalcount  (Status : Not Started ❌)

Retrieves the total count of emails. Returns a JSON object with a single field, `count`, containing the count.

### GET /search  (Status : Not Started ❌)

Searches for emails based on certain criteria. Accepts the following query parameters:

- `recipient`: The email address to search for.
- `subject`: The subject to search for.
- `sent`: A boolean indicating whether to search for emails that have been sent (`true`) or not sent (`false`).
- `opened`: A boolean indicating whether to search for emails that have been opened (`true`) or not opened (`false`).

## Email Template  (Status : Status : Completed ☑️)

The email content is generated from a Go template file, `email_template.tmpl`. This file defines two templates, `subject` and `plainBody`, which are used to generate the subject and body of the email, respectively. The templates have access to the following data:

- `Subject`: The subject of the email.
- `Body`: The body of the email.
- `Recipient`: The email address to send the email to.

The `plainBody` template also includes a tracking pixel, which calls the `/trackemail` endpoint when the email is opened.
