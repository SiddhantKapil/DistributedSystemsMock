# RESTful APIs Demo

### About This Project

My motive behind this project is to build an easy-to-use demo for showcasing how RESTful APIs work. I followed a pure RESTful approach in designing this demo and I will keep updating it with other distributed systems features. I took PDA processor as an example problem and designed microservices to manage the PDA.

### Features Covered Till now

* Handling HTTP Requests using Golang
* Setting up localhost and listen on a port
* Call APIs using curl
* Caching the data
* Write Responses

### Features Yet to Be Covered
* Mobility and Replication for PDA processors on distributed systems
* Showcase difference between Asynchronous and Synchronous calls



Create a PDA:
curl.exe -H "Content-Type: application/json" -X POST -d '{\"name\": \"HelloPDA\",\"states\": [\"q1\", \"q2\", \"q3\", \"q4\"],\"input_alphabet\": [\"0\", \"1\"],\"stack_alphabet\" : [\"0\", \"1\"],\"accepting_states\": [\"q1\", \"q4\"],\"start_state\": \"q1\",\"transitions\": [[\"q1\", null, null, \"q2\", \"$\"],[\"q2\", \"0\", null, \"q2\", \"0\"],[\"q2\", \"1\", \"0\", \"q3\", null],[\"q3\", \"1\", \"0\", \"q3\", null],[\"q3\", null, \"$\", \"q4\", null]],\"eos\": \"$\"}'  http://localhost:8888/base/pdas/198

GET PDAS:
curl.exe  http://localhost:8888/base/pdas
