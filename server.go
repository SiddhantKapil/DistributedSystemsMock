package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

type PDA struct {
	Name             string `json:"name"`
	States           []string
	Input_alphabet   []string
	Stack_alphabet   []string
	Start_state      string
	Accepting_states []string
	Transitions      [][]string
	Eos              string
	ReplicaID        string
}

type Stack []string

// IsEmpty: check if stack is empty
func (s *Stack) IsEmpty() bool {
	return len(*s) == 0
}

func (s *Stack) Push(str string) {
	*s = append(*s, str)
}

func (s *Stack) Pop() (string, bool) {
	if s.IsEmpty() {
		return "", false
	} else {
		index := len(*s) - 1
		element := (*s)[index]
		*s = (*s)[:index]
		return element, true
	}
}

var pdas = cache.New(5*time.Minute, 10*time.Minute)
var stacks = cache.New(5*time.Minute, 10*time.Minute)
var currentStates = cache.New(5*time.Minute, 10*time.Minute)
var counts = cache.New(5*time.Minute, 10*time.Minute)
var tokenQueues = cache.New(5*time.Minute, 10*time.Minute)

var replicas = cache.New(5*time.Minute, 10*time.Minute)
var sessionTokens = cache.New(5*time.Minute, 10*time.Minute)

func CreateStack(id string) {
	var stack Stack
	stack.Push("$")
	stacks.Set(id, stack, cache.NoExpiration)

	var count int
	count = 0
	counts.Set(id, count, cache.NoExpiration)

	queue := make(map[int]string)
	tokenQueues.Set(id, queue, cache.NoExpiration)

	inter, _ := pdas.Get(id)
	pda := inter.(PDA)

	current_state := pda.States[1]
	currentStates.Set(id, current_state, cache.NoExpiration)

}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func CreatePDA(id string, pda PDA, replica_id string) {
	// pda.ReplicaID = replica_id

	if replica_id != "nil" {

		inter, found := replicas.Get(replica_id)

		if found {
			replica := inter.(REPLICA)

			members := replica.Members
			_, found := Find(members, id)
			if !(found) {
				members = append(members, id)
			}

			replica.Members = members
			replicas.Set(replica_id, replica, cache.NoExpiration)
		} else {
			var replica REPLICA
			var members []string
			members = append(members, id)
			replica.Members = members
			replica.Specification = pda
			// li = append(li, id)
			replicas.Set(replica_id, replica, cache.NoExpiration)
		}
		pda.ReplicaID = replica_id
		pdas.Set(id, pda, cache.NoExpiration)
	} else {
		if pda.ReplicaID == "nil" {
			pdas.Set(id, pda, cache.NoExpiration)
		}
	}

	CreateStack(id)
}

func Open(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	var pda PDA
	inter, found := pdas.Get(id)
	var createNew bool
	createNew = true
	if found {
		pda = inter.(PDA)
		if pda.ReplicaID != "nil" {
			createNew = false
		}
	}

	if createNew {
		r.Close = true
		defer r.Body.Close()

		byteValue, err := ioutil.ReadAll(r.Body)
		json.Unmarshal(byteValue, &pda)

		// fmt.Println(string(byteValue))

		// fmt.Println(err2)
		if err != nil {
			panic(err)
		}

		err1 := json.Unmarshal([]byte(byteValue), &pda)
		if err1 != nil {
			fmt.Fprint(w, err1)
			os.Exit(1)
		}

		CreatePDA(id, pda, "nil")
	}
	fmt.Fprint(w, "PDA Created: ", id)

}

func GetPdas(w http.ResponseWriter, r *http.Request) {
	var keys []string
	for k := range pdas.Items() {
		keys = append(keys, k)
	}
	fmt.Fprint(w, keys)
}

func Reset(w http.ResponseWriter, r *http.Request) {
	fmt.Print(r.Cookies())
	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	inter, _ := pdas.Get(id)
	pda := inter.(PDA)

	sessionToken, cookieFound := r.Cookie(pda.ReplicaID)
	if cookieFound != nil {
		fmt.Fprintln(w, "Cookie not Found. Please make a connect Call first")
	} else {

		inter, _ := sessionTokens.Get(sessionToken.Value)
		cookie := inter.(COOKIE)
		cookie.CallConnectRequired = true
		cookie.Member = ""
		cookie.LastPDA = ""

		sessionTokens.Set(pda.ReplicaID, cookie, cache.NoExpiration)
		_, found := pdas.Get(id)
		if !found {
			fmt.Fprint(w, "PDA Does Not Exists!")

		} else {
			CreateStack(id)
			out, _ := json.Marshal(cookie)
			fmt.Println(string(out))
			fmt.Fprintln(w, string(out))
		}

	}
}

func InsertToken(id string, pda PDA, token string, position int) int {

	csInter, _ := currentStates.Get(id)
	currentState := csInter.(string)
	// var stack Stack
	inter, _ := stacks.Get(id)
	stack := inter.(Stack)
	var retVal int
	retVal = -1

	for i := range pda.Transitions {

		var x, y, q1, q2, z string
		var val string

		q1 = pda.Transitions[i][0]
		q2 = pda.Transitions[i][3]
		x = pda.Transitions[i][1]
		y = pda.Transitions[i][2]
		z = pda.Transitions[i][4]

		if (x == token) && (currentState == q1) {
			if y == "" {
				stack.Push(z)
				currentState = q2
				retVal = i
				break
			} else {
				if stack != nil {
					val, _ = stack.Pop()

					if val == y {
						currentState = q2
						retVal = i
						break
					} else {
						retVal = -1
						break
					}

				} else {
					retVal = -1
					break
				}

			}
		}

	}

	if retVal != -1 {
		stacks.Set(id, stack, cache.NoExpiration)
		currentStates.Set(id, currentState, cache.NoExpiration)
	}

	return retVal
}

func SyncPDAs(id1 string, id2 string) {

	// PDA1 variables
	stackInter1, _ := stacks.Get(id1)
	stack1 := stackInter1.(Stack)

	CurrentStateInter1, _ := currentStates.Get(id1)
	currentState1 := CurrentStateInter1.(string)

	countInter1, _ := counts.Get(id1)
	count1 := countInter1.(int)

	tokenQueueInter1, _ := tokenQueues.Get(id1)
	tokenQueue1 := tokenQueueInter1.(map[int]string)

	// PDA2 Variables
	stackInter2, _ := stacks.Get(id2)
	stack2 := stackInter2.(Stack)

	CurrentStateInter2, _ := currentStates.Get(id2)
	currentState2 := CurrentStateInter2.(string)

	countInter2, _ := counts.Get(id2)
	count2 := countInter2.(int)

	tokenQueueInter2, _ := tokenQueues.Get(id2)
	tokenQueue2 := tokenQueueInter2.(map[int]string)

	// Set PDA1 variables with pda2

	stack2 = stack1
	currentState2 = currentState1
	tokenQueue2 = tokenQueue1
	count2 = count1

	// Update PDA1 variables
	tokenQueues.Set(id2, tokenQueue2, cache.NoExpiration)
	stacks.Set(id2, stack2, cache.NoExpiration)
	currentStates.Set(id2, currentState2, cache.NoExpiration)
	counts.Set(id2, count2, cache.NoExpiration)
}

func Put(w http.ResponseWriter, r *http.Request) {
	fmt.Print(r.Cookies())
	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	inter, found := pdas.Get(id)
	pda := inter.(PDA)

	sessionToken, cookieFound := r.Cookie(pda.ReplicaID)
	if cookieFound != nil {
		fmt.Fprintln(w, "Cookie not Found. Please make a connect Call first")
	} else {
		inter, _ := sessionTokens.Get(sessionToken.Value)
		cookie := inter.(COOKIE)

		if cookie.CallConnectRequired {

			fmt.Fprintln(w, "Connect is not called")
		} else {

			token := strings.Split(url, "/")[4]
			posString := strings.Split(url, "/")[5]

			if cookie.Member != id {

				fmt.Println(w, "Client is not connected to this member. Try calling Member: ", cookie.Member)
			} else {

				if (cookie.LastPDA != id) && (cookie.LastPDA != "") {
					SyncPDAs(cookie.LastPDA, id)
				}
				var position int
				if n, err := strconv.Atoi(posString); err == nil {
					position = n
				} else {
					fmt.Fprint(w, n, "is not an integer.")
				}

				// json.Unmarshal(data, &pda)
				if !found {
					fmt.Fprint(w, "PDA Does Not Exists!")

				} else {

					// var reject bool
					var status int
					// reject = false
					countInter, _ := counts.Get(id)
					count := countInter.(int)

					queueInter, _ := tokenQueues.Get(id)
					queue := queueInter.(map[int]string)
					if position <= count {

					}
					if token == "eos" {
						queue[position] = pda.Eos
					} else {
						queue[position] = token
					}

					tokenQueues.Set(id, queue, cache.NoExpiration)

					for true {
						if token, ok := queue[count+1]; ok {
							if token == pda.Eos {
								inter, _ := stacks.Get(id)
								stack := inter.(Stack)
								val, _ := stack.Pop()
								if val == pda.Eos {
									fmt.Fprint(w, "Language Accepted.")

									current_state := pda.Accepting_states[1]
									currentStates.Set(id, current_state, cache.NoExpiration)
								} else {
									fmt.Fprint(w, "PDA cannot process this language.")

								}
								count++
								counts.Set(id, count, cache.NoExpiration)
								break
							}
							status = InsertToken(id, pda, token, position)
							count++

							if status == -1 {
								fmt.Fprint(w, "PDA cannot process this language.")

								break
							} else {
								fmt.Fprintln(w, "Transition done. Status: ", status)
								counts.Set(id, count, cache.NoExpiration)
							}
						} else {
							break
						}

					}

					cookie.LastPDA = id
					cookie.CallConnectRequired = true
					sessionTokens.Set(sessionToken.Value, cookie, cache.NoExpiration)

				}
			}
		}
	}
}

func peek(k int, stack Stack) []string {

	if k == 0 {
		k = 1
	}

	index := len(stack) - k
	if index <= 0 {
		return (stack)[:]
	} else {
		return (stack)[index:]
	}
}

func TopK(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	inter, _ := pdas.Get(id)
	pda1 := inter.(PDA)

	sessionToken, cookieFound := r.Cookie(pda1.ReplicaID)

	if cookieFound != nil {
		fmt.Fprintln(w, "Cookie not Found. Please make a connect Call first")
	} else {

		intercookie, _ := sessionTokens.Get(sessionToken.Value)
		cookie := intercookie.(COOKIE)

		id2 := cookie.LastPDA
		SyncPDAs(id2, id)

		url := r.URL.Path
		// id := strings.Split(url, "/")[3]
		kString := strings.Split(url, "/")[6]
		var k int

		if n, err := strconv.Atoi(kString); err == nil {
			k = n
		} else {
			fmt.Fprint(w, n, "is not an integer.")
		}

		inter, found := stacks.Get(id)

		// json.Unmarshal(data, &pda)
		if !found {
			fmt.Fprintf(w, "PDA Does Not Exists!")

		} else {

			stack := inter.(Stack)
			output := peek(k, stack)
			fmt.Fprint(w, output)
		}
	}
}

func isAccepted(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	inter, _ := pdas.Get(id)
	pda1 := inter.(PDA)

	sessionToken, cookieFound := r.Cookie(pda1.ReplicaID)

	if cookieFound != nil {
		fmt.Fprintln(w, "Cookie not Found. Please make a connect Call first")
	} else {

		intercookie, _ := sessionTokens.Get(sessionToken.Value)
		cookie := intercookie.(COOKIE)

		id2 := cookie.LastPDA
		SyncPDAs(id2, id)

		// url := r.URL.Path
		// id := strings.Split(url, "/")[3]

		inter, found := pdas.Get(id)

		// json.Unmarshal(data, &pda)
		if !found {
			fmt.Fprint(w, "PDA Does Not Exists!")

		} else {

			csInter, _ := currentStates.Get(id)
			currentState := csInter.(string)
			pda := inter.(PDA)
			var isAccept bool
			isAccept = false
			for i := range pda.Accepting_states {
				if currentState == pda.Accepting_states[i] {
					isAccept = true
				}
			}

			fmt.Fprint(w, isAccept)
		}
	}
}

func StackLength(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	inter, _ := pdas.Get(id)
	pda1 := inter.(PDA)

	sessionToken, cookieFound := r.Cookie(pda1.ReplicaID)

	if cookieFound != nil {

		fmt.Fprintln(w, "Cookie not Found. Please make a connect Call first")
	} else {

		intercookie, _ := sessionTokens.Get(sessionToken.Value)
		cookie := intercookie.(COOKIE)

		id2 := cookie.LastPDA
		SyncPDAs(id2, id)

		// url := r.URL.Path
		// id := strings.Split(url, "/")[3]

		inter, found := stacks.Get(id)

		if !found {
			fmt.Fprint(w, "PDA Does Not Exists!")
		} else {

			stack := inter.(Stack)
			fmt.Fprint(w, len(stack))

		}
	}
}

func CurrentState(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	inter, _ := pdas.Get(id)
	pda1 := inter.(PDA)

	sessionToken, cookieFound := r.Cookie(pda1.ReplicaID)

	if cookieFound != nil {
		fmt.Fprintln(w, "Cookie not Found. Please make a connect Call first")
	} else {

		intercookie, _ := sessionTokens.Get(sessionToken.Value)
		cookie := intercookie.(COOKIE)

		id2 := cookie.LastPDA
		SyncPDAs(id2, id)

		// url := r.URL.Path
		// id := strings.Split(url, "/")[3]
		inter, found := currentStates.Get(id)

		if !found {
			fmt.Fprint(w, "PDA Does Not Exists!")
		} else {

			currentState := inter.(string)
			fmt.Fprint(w, currentState)
		}
	}

}

func QueuedTokens(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	inter, _ := pdas.Get(id)
	pda1 := inter.(PDA)

	sessionToken, cookieFound := r.Cookie(pda1.ReplicaID)

	if cookieFound != nil {
		fmt.Fprintln(w, "Cookie not Found. Please make a connect Call first")
	} else {

		intercookie, _ := sessionTokens.Get(sessionToken.Value)
		cookie := intercookie.(COOKIE)

		id2 := cookie.LastPDA
		SyncPDAs(id2, id)

		// url := r.URL.Path
		// id := strings.Split(url, "/")[3]
		inter, found := tokenQueues.Get(id)

		if !found {
			fmt.Fprintf(w, "PDA Does Not Exists!")
		} else {

			queue := inter.(map[int]string)
			fmt.Fprint(w, queue)
		}
	}
}

func SnapShot(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	inter, _ := pdas.Get(id)
	pda1 := inter.(PDA)

	sessionToken, cookieFound := r.Cookie(pda1.ReplicaID)

	if cookieFound != nil {
		fmt.Fprintln(w, "Cookie not Found. Please make a connect Call first")
	} else {

		intercookie, _ := sessionTokens.Get(sessionToken.Value)
		cookie := intercookie.(COOKIE)

		id2 := cookie.LastPDA
		SyncPDAs(id2, id)
		// id := strings.Split(url, "/")[3]
		kString := strings.Split(url, "/")[5]
		var k int
		if n, err := strconv.Atoi(kString); err == nil {
			k = n
		} else {
			fmt.Fprint(w, kString, "is not an integer.")
		}

		inter, found := tokenQueues.Get(id)

		if !found {
			fmt.Fprint(w, "PDA Does Not Exists!")
		} else {

			queue := inter.(map[int]string)
			fmt.Fprintln(w, "Token Queue: ", queue)

			interStack, _ := stacks.Get(id)
			stack := interStack.(Stack)
			output := peek(k, stack)
			fmt.Fprintln(w, "Peek ", k, ": ", output)

			inter, _ := currentStates.Get(id)
			currentState := inter.(string)
			fmt.Fprintln(w, "Current State: ", currentState)
		}
	}
}

func Close(w http.ResponseWriter, r *http.Request) {

}
func Delete(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]

	_, found := pdas.Get(id)

	// json.Unmarshal(data, &pda)
	if !found {
		fmt.Fprint(w, "PDA Does Not Exists!")

	} else {

		currentStates.Delete(id)
		pdas.Delete(id)
		tokenQueues.Delete(id)
		counts.Delete(id)
		stacks.Delete(id)

	}
}

type REPLICA struct {
	Members       []string
	Specification PDA
}

func CloseReplica(w http.ResponseWriter, r *http.Request) {

}

func DeleteReplica(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]

	inter, found := replicas.Get(id)

	// json.Unmarshal(data, &pda)
	if !found {
		fmt.Fprint(w, "PDA Does Not Exists!")

	} else {

		replica := inter.(REPLICA)
		for _, member := range replica.Members {
			fmt.Fprintln(w, "Deleting Member:", member)
			currentStates.Delete(member)
			pdas.Delete(member)
			tokenQueues.Delete(member)
			counts.Delete(member)
			stacks.Delete(member)

		}
		replicas.Delete(id)
	}
}

func CreateReplicaGroup(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	var replica REPLICA
	_, found := replicas.Get(id)

	if found {
		fmt.Fprint(w, "Replica Group Already Exists!")
	} else {
		r.Close = true
		defer r.Body.Close()

		byteValue, err := ioutil.ReadAll(r.Body)
		json.Unmarshal(byteValue, &replica)

		// fmt.Println(string(byteValue))

		// fmt.Println(err2)
		if err != nil {
			panic(err)
		}
		replicas.Set(id, replica, cache.NoExpiration)
		for _, member := range replica.Members {
			fmt.Fprintln(w, "Update Specs for Member:", member)
			CreatePDA(member, replica.Specification, id)

		}
		fmt.Fprint(w, "Replica Created: ", id)
	}
}

func GetReplicas(w http.ResponseWriter, r *http.Request) {
	var keys []string
	for k := range replicas.Items() {
		keys = append(keys, k)
	}
	fmt.Fprint(w, "Replicas: ", keys)
}

func ResetReplica(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]

	inter, found := replicas.Get(id)
	replica := inter.(REPLICA)
	if !found {
		fmt.Fprint(w, "Replica Does Not Exists!")

	} else {
		for _, member := range replica.Members {
			fmt.Fprint(w, "Resetting Member: ", member)
			CreateStack(member)
		}

	}
}

func MembersReplica(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]

	inter, found := replicas.Get(id)
	replica := inter.(REPLICA)
	if !found {
		fmt.Fprint(w, "Replica Does Not Exists!")

	} else {
		fmt.Fprintln(w, "members in Replica", id)
		for _, member := range replica.Members {
			fmt.Fprintln(w, member)
		}

	}
}

type COOKIE struct {
	ReplicaID           string
	LastPDA             string
	Member              string
	CallConnectRequired bool
}

func Connect(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]

	inter, found := replicas.Get(id)
	replica := inter.(REPLICA)
	if !found {
		fmt.Fprint(w, "Replica Does Not Exists!")

	} else {

		rand.Seed(time.Now().Unix())
		members := replica.Members
		n := rand.Int() % len(members)

		// Check for Cookies
		cookieObtained, cookieFound := r.Cookie(id)
		var sessionToken string

		if cookieFound != nil {
			var value COOKIE
			value.CallConnectRequired = false
			value.Member = members[n]
			value.ReplicaID = id
			value.LastPDA = ""

			// Add Cookie
			sessionToken = uuid.New().String()
			sessionTokens.Set(sessionToken, value, cache.NoExpiration)

		} else {
			interToken, _ := sessionTokens.Get(cookieObtained.Value)
			value := interToken.(COOKIE)
			value.CallConnectRequired = false
			value.Member = members[n]
			sessionToken = cookieObtained.Value
			sessionTokens.Set(sessionToken, value, cache.NoExpiration)
		}

		expire := time.Now().AddDate(0, 0, 1)
		cookie := http.Cookie{
			Name:    id,
			Value:   sessionToken,
			Expires: expire,
		}

		http.SetCookie(w, &cookie)

		fmt.Fprint(w, members[n])
	}
}

type JOIN struct {
	Replica_id string
}

func Join(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]

	r.Close = true
	defer r.Body.Close()
	var join JOIN
	byteValue, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(byteValue, &join)

	// fmt.Println(string(byteValue))

	inter, found := replicas.Get(join.Replica_id)

	if !found {
		fmt.Fprint(w, "Replica Does Not Exists!")

	} else {

		replica := inter.(REPLICA)
		inter1, _ := pdas.Get(id)
		pda := inter1.(PDA)
		if join.Replica_id != pda.ReplicaID {
			CreatePDA(id, replica.Specification, join.Replica_id)
			fmt.Fprint(w, "Pda ", id, " Successully Joined")
		} else {
			fmt.Fprint(w, "Pda Already Exists in the Replica")
		}
	}
}

func Code(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]

	// fmt.Println(string(byteValue))

	inter, found := pdas.Get(id)

	if !found {
		fmt.Fprint(w, "PDA  Does Not Exists!")

	} else {

		pda := inter.(PDA)
		out, _ := json.Marshal(pda)
		fmt.Fprintln(w, string(out))
	}
}

func C3State(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	inter, _ := pdas.Get(id)
	pda1 := inter.(PDA)

	sessionToken, cookieFound := r.Cookie(pda1.ReplicaID)

	if cookieFound != nil {
		fmt.Fprintln(w, "Cookie not Found. Please make a connect Call first")
	} else {

		intercookie, _ := sessionTokens.Get(sessionToken.Value)
		cookie := intercookie.(COOKIE)
		out, _ := json.Marshal(cookie)
		fmt.Println(w, string(out))
	}
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/base/pdas", GetPdas)
	r.HandleFunc("/base/pdas/{id}", Open)
	r.HandleFunc("/base/pdas/{id}/reset", Reset)
	r.HandleFunc("/base/pdas/{id}/snapshot/{k}", SnapShot)
	r.HandleFunc("/base/pdas/{id}/{token}/{position}", Put)
	r.HandleFunc("/base/pdas/{id}/is_accepted", isAccepted)
	r.HandleFunc("/base/pdas/{id}/stack/top/{k}", TopK)
	r.HandleFunc("/base/pdas/{id}/stack/len", StackLength)
	r.HandleFunc("/base/pdas/{id}/state", CurrentState)
	r.HandleFunc("/base/pdas/{id}/tokens", QueuedTokens)
	r.HandleFunc("/base/pdas/{id}/close", Close)
	r.HandleFunc("/base/pdas/{id}/delete", Delete)
	r.HandleFunc("/base/pdas/{id}/join", Join)
	r.HandleFunc("/base/pdas/{id}/code", Code)
	r.HandleFunc("/base/pdas/{id}/c3state", C3State)

	r.HandleFunc("/base/replica_pdas", GetReplicas)
	r.HandleFunc("/base/replica_pdas/{gid}", CreateReplicaGroup)
	r.HandleFunc("/base/replica_pdas/{gid}/reset", ResetReplica)
	r.HandleFunc("/base/replica_pdas/{gid}/members", MembersReplica)
	r.HandleFunc("/base/replica_pdas/{gid}/connect", Connect)
	r.HandleFunc("/base/replica_pdas/{gid}/close", CloseReplica)
	r.HandleFunc("/base/replica_pdas/{gid}/delete", DeleteReplica)

	http.ListenAndServe(":8080", r)
}
