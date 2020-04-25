Project Members:

Pratiksha Patnaik
Siddhant Raj Kapil
Charu Sharma





Create a PDA:
curl.exe -H "Content-Type: application/json" -X POST -d '{\"name\": \"HelloPDA\",\"states\": [\"q1\", \"q2\", \"q3\", \"q4\"],\"input_alphabet\": [\"0\", \"1\"],\"stack_alphabet\" : [\"0\", \"1\"],\"accepting_states\": [\"q1\", \"q4\"],\"start_state\": \"q1\",\"transitions\": [[\"q1\", null, null, \"q2\", \"$\"],[\"q2\", \"0\", null, \"q2\", \"0\"],[\"q2\", \"1\", \"0\", \"q3\", null],[\"q3\", \"1\", \"0\", \"q3\", null],[\"q3\", null, \"$\", \"q4\", null]],\"eos\": \"$\"}'  http://localhost:8888/base/pdas/198

GET PDAS:
curl.exe  http://localhost:8888/base/pdas