
## Prompts

1. Create the boilerplate for a Go REST API. It should use the Gin framework for requests and have extensible utilities for making all HTTP requests that can be applied to different endpoints.
2. Create a package /protos and add proto files with the following message definitions:
- Item, which has the following fields:
    - A unique identifier
    - A url link (optional)
    - A thumbnail photo link (optional)
    - Notes (as a string)
- Wishlist, which has the following fields:
    - A unique identifier
    - An array of Items
