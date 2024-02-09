# Slug
<img align="right" width="159px" src="https://i.ibb.co/p2zvQ36/DALL-E-2024-02-08-16-19-35-Animated-Slug-mascot-in-teal-color-racing-in-blazing-speed-showing-full-i.png">

 An experimental static file server using a custom HTTP framework built as a part of assignment for Distributed Systems (CSEN 371) at Santa Clara University.

 ## Features

  - The HTTP framework has been built from the groundup, focusing on modularity, customisability and extendibility. It features a router-handler style library which can be imported to build REST APIs. 
  - Currently only supports HTTP/1.0 and HTTP/1.1 protocols accroding to the project requirements.
  - Each request is handled by a seperate Goroutine lightweight thread.
  - Gracefully shuts down the server when OS signals like _SIGINT_ & _SIGTERM_ are caught. The maximum amount of time taken by the server to wait for pending requests is configurable.
  - Uses heuristics to determine the timeout for `keep-alive` based on number of active connections.

## Usage

Make sure `go` is installed and set in the path.
```
$ brew install go
```
To build the project locally, 
```
$ make build_local
```
An executable will be generated in `bin/` directory. To run it,
```
# run slug server and visit localhost:8080 on browser
$ ./bin/slug -document_root=/Users/www/scu.edu -port=8080 -timeout=5
```
<img align="left" width="550px" src="https://github.com/roychowdhuryrohit-dev/slug/assets/24897721/1e0574e8-e2fc-4064-b950-a7fd7ac21181">


