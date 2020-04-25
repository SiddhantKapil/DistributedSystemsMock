 #!/bin/bash
echo 'Creating PDA with id: 199'
curl -H "Content-Type: application/json" -X POST -d \
'{    "name": "HelloPDA",    "states": ["q1", "q2", "q3", "q4"],    "input_alphabet": ["0", "1"],    "stack_alphabet" : ["0", "1"],    "accepting_states": ["q1", "q4"],    "start_state": "q1",    "transitions": [    ["q1", null, null, "q2", "$"],    ["q2", "0", null, "q2", "0"],    ["q2", "1", "0", "q3", null],    ["q3", "1", "0", "q3", null],    ["q3", null, "$", "q4", null]],    "eos": "$"}'  http://localhost:8080/base/pdas/199
echo 'input token 0 at position 2'
curl  http://localhost:8080/base/pdas/199/0/2
echo 'input token 0 at position 1'
curl  http://localhost:8080/base/pdas/199/0/1
echo 'input token 0 at position 3'
curl  http://localhost:8080/base/pdas/199/0/3
echo 'Snapshot with k=4'
curl  http://localhost:8080/base/pdas/199/snapshot/4
echo 'Call is_accepted'
curl  http://localhost:8080/base/pdas/199/is_accepted
echo 'input token 1 at position 5'
curl  http://localhost:8080/base/pdas/199/1/5
echo 'input token 1 at position 4'
curl  http://localhost:8080/base/pdas/199/1/4
echo 'input token 1 at position 6'
curl  http://localhost:8080/base/pdas/199/1/6
echo 'input eos at position 7' 
curl  http://localhost:8080/base/pdas/199/eos/7
echo 'Call is_accepted'
curl  http://localhost:8080/base/pdas/199/is_accepted
echo 'Call Reset Stack'
curl  http://localhost:8080/base/pdas/199/reset
