import requests
import json

#  cookies = {"enwiki_session": "17ab96bd8ffbe8ca58a78657a918558"}


# make two specifications

# specification 1 for language 0n1n
specification1 = """{"name": "HelloPDA",
"states": ["q1", "q2", "q3", "q4"],
"input_alphabet": ["0", "1"], 
"stack_alphabet" : ["0", "1"], 
"accepting_states": ["q1", "q4"],
"replicaID": "nil",
"start_state": "q1", 
"transitions": [["q1", null, null, "q2", "$"],
["q2", "0", null, "q2", "0"],
 ["q2", "1", "0", "q3", null],
  ["q3", "1", "0", "q3", null], 
  ["q3", null, "$", "q4", null]],
 "eos": "$"}"""

# specification 2 for language 1n0n
specification2 = """{"name": "HelloPDA",
    "states": ["q1", "q2", "q3", "q4"],
    "input_alphabet": ["0", "1"],
    "stack_alphabet" : ["0", "1"],
    "accepting_states": ["q1", "q4"],
    "start_state": "q1",
    "replicaID": "nil",
    "transitions": [
    ["q1", null, null, "q2", "$"],
    ["q2", "1", null, "q2", "1"],
    ["q2", "0", "1", "q3", null],
    ["q3", "0", "1", "q3", null],
    ["q3", null, "$", "q4", null]],
    "eos": "$"}"""

content1 = json.loads(specification1)
content2 = json.loads(specification2)


# create pdas with specification 1 
r = requests.post("http://localhost:8080/base/pdas/101", json=content1)
print(r.text)
r = requests.post("http://localhost:8080/base/pdas/102", json=content1)
print(r.text)
r = requests.post("http://localhost:8080/base/pdas/103", json=content1)
print(r.text)
r = requests.post("http://localhost:8080/base/pdas/104", json=content1)
print(r.text)
print()

# Create pdas with specification 2
r = requests.post("http://localhost:8080/base/pdas/201", json=content2) 
print(r.text)
r = requests.post("http://localhost:8080/base/pdas/202", json=content2) 
print(r.text)
r = requests.post("http://localhost:8080/base/pdas/203", json=content2)
print(r.text)
r = requests.post("http://localhost:8080/base/pdas/204", json=content2)
print(r.text)
print()

###########################
# Create Replica 1 config #
###########################

# create specification for making a replica group
replica1_specs = """{
    "name": "HelloPDA",
    "states": ["q1", "q2", "q3", "q4"],
    "input_alphabet": ["0", "1"], 
    "stack_alphabet" : ["0", "1"], 
    "accepting_states": ["q1", "q4"],
    "start_state": "q1",
    "replicaID": "nil",
    "transitions": [["q1", null, null, "q2", "$"],
    ["q2", "0", null, "q2", "0"],
    ["q2", "1", "0", "q3", null],
    ["q3", "1", "0", "q3", null], 
     ["q3", null, "$", "q4", null]],
    "eos": "$"
    }"""
# Convert replica specification to json
replica_content1 = json.loads(replica1_specs)

# Create json for replica 1
replica1_configs = {
    "members": ["101", "102", "104"],
    "specification" : replica_content1

 }

###########################
# Create Replica 2 config #
###########################

# create specification for making a replica group
replica2_specs = """{"name": "HelloPDA",
    "states": ["q1", "q2", "q3", "q4"],
    "input_alphabet": ["0", "1"],
    "stack_alphabet" : ["0", "1"],
    "accepting_states": ["q1", "q4"],
    "start_state": "q1",
    "replicaID": "nil",
    "transitions": [
    ["q1", null, null, "q2", "$"],
    ["q2", "1", null, "q2", "1"],
    ["q2", "0", "1", "q3", null],
    ["q3", "0", "1", "q3", null],
    ["q3", null, "$", "q4", null]],
    "eos": "$"}"""

# Convert replica specification to json
replica_content2 = json.loads(replica2_specs)

# Create json for replica 1
replica2_configs = {
    "members": ["201", "202", "203"],
    "specification" : replica_content2

 }

###########################
# Create Replica 1 and 2  #
###########################

# Create replica 1 group with tw members 101 and 102 and specification as shown above
r = requests.post("http://localhost:8080/base/replica_pdas/1", json=replica1_configs) 
print("http://localhost:8080/base/replica_pdas/1")
print(r.text)
print()

# Create replica 2 group with tw members 201 and 202 and specification as shown above
r = requests.post("http://localhost:8080/base/replica_pdas/2", json=replica2_configs) 
print("http://localhost:8080/base/replica_pdas/2")
print(r.text)
print()

# Get Replicas
r = requests.get("http://localhost:8080/base/replica_pdas") 
print("http://localhost:8080/base/replica_pdas")
print( r.text)
print()

# Join Configs for replica 1 and 2
join1_config = {
    "replica_id": "1"
}
join2_config = {
    "replica_id": "2"
}

# join 103 in replica 1
r = requests.post("http://localhost:8080/base/pdas/103/join", json=join1_config) 
print("http://localhost:8080/base/pdas/103/join")
print(r.text)
print()

# join 204 in replica 2
r = requests.post("http://localhost:8080/base/pdas/204/join", json=join2_config) 
print("http://localhost:8080/base/pdas/204/join")
print(r.text)
print()

r = requests.get("http://localhost:8080/base/replica_pdas/1/members") 
print("http://localhost:8080/base/replica_pdas/1/members")
print(r.text)
print()

print("#############################################################################################")
print("Feeding 000111$ to Replica 1 PDAs and 1100$ to PDA2 as PDA1 accept 0n1n and PDA2 accepts 1n0n")
print("#############################################################################################")
print()
# Create a session
session = requests.Session()
print("Created a New Session")
print("Cookies: ", session.cookies.get_dict())

# For Replica 1
r = session.get("http://localhost:8080/base/replica_pdas/1/connect", cookies = session.cookies.get_dict()) 
print("http://localhost:8080/base/replica_pdas/1/connect")
print("Connecting to member: ", r.text)

r = session.post("http://localhost:8080/base/pdas/"+r.text+"/0/1", cookies = session.cookies.get_dict())
print(r.text)

r = session.get("http://localhost:8080/base/replica_pdas/1/connect", cookies = session.cookies.get_dict()) 
print("http://localhost:8080/base/replica_pdas/1/connect")
print("Connecting to member: ", r.text)

r = session.post("http://localhost:8080/base/pdas/"+r.text+"/0/3", cookies = session.cookies.get_dict())
print(r.text)

r = session.get("http://localhost:8080/base/replica_pdas/1/connect", cookies = session.cookies.get_dict()) 
print("http://localhost:8080/base/replica_pdas/1/connect")
print("Connecting to member: ", r.text)

r = session.post("http://localhost:8080/base/pdas/"+r.text+"/1/5", cookies = session.cookies.get_dict())
print(r.text)

r = session.get("http://localhost:8080/base/replica_pdas/1/connect", cookies = session.cookies.get_dict()) 
print("http://localhost:8080/base/replica_pdas/1/connect")
print("Connecting to member: ", r.text)

r = session.post("http://localhost:8080/base/pdas/"+r.text+"/0/2", cookies = session.cookies.get_dict())
print(r.text)

r = session.get("http://localhost:8080/base/replica_pdas/1/connect", cookies = session.cookies.get_dict()) 
print("http://localhost:8080/base/replica_pdas/1/connect")
print("Connecting to member: ", r.text)

r = session.post("http://localhost:8080/base/pdas/"+r.text+"/1/6", cookies = session.cookies.get_dict())
print(r.text)

# call snapshot
r = session.post("http://localhost:8080/base/pdas/102/snapshot/4", cookies = session.cookies.get_dict())
print("Calling snapshot for Replica 1")
print("http://localhost:8080/base/pdas/102/snapshot/4")
print(r.text)
print()

r = session.get("http://localhost:8080/base/replica_pdas/1/connect", cookies = session.cookies.get_dict()) 
print("http://localhost:8080/base/replica_pdas/1/connect")
print("Connecting to member: ", r.text)

r = session.post("http://localhost:8080/base/pdas/"+r.text+"/1/4", cookies = session.cookies.get_dict())
print(r.text)
print()

r = session.post("http://localhost:8080/base/pdas/103/is_accepted", cookies = session.cookies.get_dict())
print("http://localhost:8080/base/pdas/103/is_accepted")
print("Calling Is Accepted for Replica 1")
print(r.text)
print()



# For replica 2
r = session.get("http://localhost:8080/base/replica_pdas/2/connect", cookies = session.cookies.get_dict()) 
print("http://localhost:8080/base/replica_pdas/2/connect")
print("Connecting to member: ", r.text)

r = session.post("http://localhost:8080/base/pdas/"+r.text+"/0/4", cookies = session.cookies.get_dict())
print(r.text)


# For replica 1
r = session.get("http://localhost:8080/base/replica_pdas/1/connect", cookies = session.cookies.get_dict()) 
print("http://localhost:8080/base/replica_pdas/1/connect")
print("Connecting to member: ", r.text)

r = session.post("http://localhost:8080/base/pdas/"+r.text+"/eos/7", cookies = session.cookies.get_dict())
print(r.text)
print()

# Call is_Accepted
r = session.post("http://localhost:8080/base/pdas/103/is_accepted", cookies = session.cookies.get_dict())
print("Calling Is Accepted for Replica 1")
print("http://localhost:8080/base/pdas/103/is_accepted")
print(r.text)
print()



#  for replica 2
r = session.get("http://localhost:8080/base/replica_pdas/2/connect", cookies = session.cookies.get_dict()) 
print("http://localhost:8080/base/replica_pdas/2/connect")
print("Connecting to member: ", r.text)

r = session.post("http://localhost:8080/base/pdas/"+r.text+"/1/1", cookies = session.cookies.get_dict())
print(r.text)

r = session.get("http://localhost:8080/base/replica_pdas/2/connect", cookies = session.cookies.get_dict()) 
print("http://localhost:8080/base/replica_pdas/2/connect")
print("Connecting to member: ", r.text)

r = session.post("http://localhost:8080/base/pdas/"+r.text+"/1/2", cookies = session.cookies.get_dict())
print(r.text)

r = session.get("http://localhost:8080/base/replica_pdas/2/connect", cookies = session.cookies.get_dict()) 
print("http://localhost:8080/base/replica_pdas/2/connect")
print("Connecting to member: ", r.text)

r = session.post("http://localhost:8080/base/pdas/"+r.text+"/0/3", cookies = session.cookies.get_dict())
print(r.text)


r = session.get("http://localhost:8080/base/replica_pdas/2/connect", cookies = session.cookies.get_dict()) 
print("http://localhost:8080/base/replica_pdas/2/connect")
print("Connecting to member: ", r.text)

r = session.post("http://localhost:8080/base/pdas/"+r.text+"/eos/5", cookies = session.cookies.get_dict())
print(r.text)
print()

# Call is accepted
r = session.post("http://localhost:8080/base/pdas/203/is_accepted", cookies = session.cookies.get_dict())
print("Calling Is Accepted for Replica 2")
print("http://localhost:8080/base/pdas/203/is_accepted")
print(r.text)
print()

# Reset Stack
r = session.post("http://localhost:8080/base/pdas/103/reset", cookies = session.cookies.get_dict())
print("http://localhost:8080/base/pdas/203/reset")
print("Calling Reset")
print(r.text)

# Call Code for 103
r = session.get("http://localhost:8080/base/pdas/103/code")
print("Calling Code for 103")
print("http://localhost:8080/base/pdas/103/code")
print(r.text)
print()

# Call Get Replicas and Members
r = session.get("http://localhost:8080/base/replica_pdas")
print("Calling replica_pdas...")
print("http://localhost:8080/base/replica_pdas")
print(r.text)
print()

r = session.get("http://localhost:8080/base/replica_pdas/1/members")
print("Calling members for replica 1")
print("http://localhost:8080/base/replica_pdas1/members")
print(r.text)
print()

r = session.get("http://localhost:8080/base/replica_pdas/2/members")
print("Calling members for replica 2")
print("http://localhost:8080/base/replica_pdas/2/members")
print(r.text)
print()

# Call Get Replicas and Members
r = session.post("http://localhost:8080/base/replica_pdas/1/delete")
print("Calling replica_pdas...")
print("http://localhost:8080/base/replica_pdas/1/delete")
print(r.text)
print()

# Call Get Replicas and Members
r = session.get("http://localhost:8080/base/replica_pdas")
print("Calling replica_pdas, after deleting replica 1")
print("http://localhost:8080/base/replica_pdas")
print(r.text)
print()