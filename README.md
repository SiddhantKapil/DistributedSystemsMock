# Distributed Systems with RESTful API demo

### About This Project

My motive behind this project is to build an easy-to-use demo for showcasing how Distributed systems work. I took a PDA as a processing server and added multiple PDAs to simulate a distributed system. Client can make several different API calls and perform CRUD operations, I took a pure RESTful approach in designing these APIs.Note this project is client-centric.

### Features Covered Till now

* Client can introduce new PDAs.
* Client can create new Replica groups by passing PDA ids in the url. Or client can join PDAs to existing Replicas.
* Client can perform CRUD operations on PDA and Replicas. All the CRUD operations are showcased in client.py.
* The states of PDAs and Replicas are cached on server side.
* To introduce mobility, client can connect to any of the PDA in a replica group and that PDA can continue client's operation by retrieving previous information using cookies.


### How To Use
* Run the server run `./server.sh`
* On a new terminal run `python3 client.py`. Client will make all the API call and you can see the outputs.

I will be glad to hear from you if this helped you in any way. Feedbacks are welcome. Thank you.
