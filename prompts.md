
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

(Used GitHub agent)
3. This API is a Wishlist-sharing service. This means the requests pertain to wishlists that track and maintain items. Create a GET request to fetch the public wishlists for a user, given some ID in the query parameters. Make sure to reuse existing REST API utilities in this application.
4. Create a POST request for a user to create a new wishlist and a DELETE request that takes the ID of wishlist in the URI to delete it by ID.
(Manual adjustments - some formatting not entirely what I had in mind.)
5. Create a PUT request for a user to update a wishlist, where the request body can take in any field value that exists on the wishlist. It should use the same URI structure as DELETE.
6. 