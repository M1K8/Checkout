# Build

`go get ./cmd`

`go build ./cmd`
Running the executable starts a server running at `localhost:1337`

# Test

Inside `internal/api`, run `go test`


# File Structure

## cmd/
### main.go
Entry point, simply serves the http server & router.

## internal/api
### auth.go
Houses the /authorize endpoint.

### capture.go
Houses the /capture endpoint.

### refund.go
Houses the /refund endpoint.

### void.go
Houses the /void endpoint.

### types.go
Contains types used across multiples files.

### utils.go
Contains methods used across multiples files.

### *_test.go
Associated unit tests for the named file.

### internal/api/info.log
This will be generated on first startup, and will contain any errors

### db/
This directory containts the embdedded DB

### testdb/
This directory contains the DB used during testing







# Assumptions made

### Card numbers & CVV
As no actual Credit Card database exists for this project, and assumption was made that any card # provided - assuming it passes a Luhn check - is valid. In order
to allow testing for a CVV mismatch, an arbritrary check was implemented that causes an error in a given circumstance.

### Card number exclusivity
The assumption was made that there can be multiple transactions with the same card number.

### Input
An assumption is made that all input is valid i.e. CC numbers are 16 digits, expiry dates are ##/##, CVV is ### etc.


# Potential improvements

### Persistent storage
In order to save time & complexity, an embedded database was used to demonstrate persistent storage. In a production environment, this would not be optimal, as it would limit the access 
of the DB to only the machine the HTTP server is running on. Viable alternatives could be a SQL derived DB.

### Concurrency
As alluded to in the above section, the persistent storage exists locally. Due to file restraints, the concurrency of access to the DB might be hindered, althugh the router and API implementations do allow for multiple endpoints to be hit concurrently.
A solution to this would be a remote DB that allows for concurrent connections.

### Authentication
For the sake of time, no authentication was added, yet in a production environment this would be highly unadvisable. For something as delicate as online payments,
something like OAuth would be the required, at the least.

### Containerisation
The modular approach taken in designing the API would allow for easy containerisation - each endpoint could exist in isolation. 
This could be beneficial, as it allows each endpoint to have a degree of autonomy, meaning maintainance on a given endpoint can be performed with minimal downtime to the other, for example. 
Due to time constraints, this was not implmented.

### Code organisation
The structure of the code, whilst fairly modularised, can be somewhat messy at time,with extensive parameter lists for some methods, and was done due to a lack of time, and to enable for easier unit testing. This can make the code somewhat difficult to maintain, can could be imporved by restructuring the flow of data.

### Test coverage
Whilst there is good code coverage through unit tests, due to time constraints, there are no embedded or integration tests. These would further increase the stability of the API by finding issues out of the scope of unit tests.
