package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

func Open(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	var pda PDA
	_, found := pdas.Get(id)
	if found {
		fmt.Fprintln(w, "PDA Already Exists!")

	} else {

		r.Close = true
		defer r.Body.Close()

		byteValue, err := ioutil.ReadAll(r.Body)
		// fmt.Println(string(byteValue))
		if err != nil {
			panic(err)
		}

		err1 := json.Unmarshal([]byte(byteValue), &pda)
		if err1 != nil {
			fmt.Fprintln(w, err1)
			os.Exit(1)
		}
		// bytePDA, _ := json.Marshal(pda)
		pdas.Set(id, pda, cache.NoExpiration)
		CreateStack(id)
	}
	// fmt.Print(pdas.Items())
}

func GetPdas(w http.ResponseWriter, r *http.Request) {
	var keys []string
	for k := range pdas.Items() {
		keys = append(keys, k)
	}
	fmt.Fprintln(w, keys)
}

func Reset(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]

	_, found := pdas.Get(id)
	if !found {
		fmt.Fprintln(w, "PDA Does Not Exists!")

	} else {
		CreateStack(id)
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

func Put(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	token := strings.Split(url, "/")[4]
	posString := strings.Split(url, "/")[5]
	var position int
	if n, err := strconv.Atoi(posString); err == nil {
		position = n
	} else {
		fmt.Fprintln(w, n, "is not an integer.")
	}

	inter, found := pdas.Get(id)

	// json.Unmarshal(data, &pda)
	if !found {
		fmt.Fprintln(w, "PDA Does Not Exists!")

	} else {

		pda := inter.(PDA)
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
						fmt.Fprintln(w, "Language Accepted.")
						current_state := pda.Accepting_states[1]
						currentStates.Set(id, current_state, cache.NoExpiration)
					} else {
						fmt.Fprintln(w, "PDA cannot process this language.")
					}

					break
				}
				status = InsertToken(id, pda, token, position)
				count++

				if status == -1 {
					fmt.Fprintln(w, "PDA cannot process this language.")

					break
				} else {
					fmt.Fprintln(w, "Transition done. Status: ", status)
					counts.Set(id, count, cache.NoExpiration)
				}
			} else {
				break
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
	kString := strings.Split(url, "/")[6]
	var k int
	if n, err := strconv.Atoi(kString); err == nil {
		k = n
	} else {
		fmt.Fprintln(w, n, "is not an integer.")
	}

	inter, found := stacks.Get(id)

	// json.Unmarshal(data, &pda)
	if !found {
		fmt.Fprintf(w, "PDA Does Not Exists!")

	} else {

		stack := inter.(Stack)
		output := peek(k, stack)
		fmt.Fprintln(w, output)
	}
}
func isAccepted(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	inter, found := pdas.Get(id)

	// json.Unmarshal(data, &pda)
	if !found {
		fmt.Fprintln(w, "PDA Does Not Exists!")

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
		fmt.Fprintln(w, isAccept)
	}
}

func StackLength(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]

	inter, found := stacks.Get(id)

	if !found {
		fmt.Fprintln(w, "PDA Does Not Exists!")
	} else {

		stack := inter.(Stack)
		fmt.Fprintln(w, len(stack))

	}
}

func CurrentState(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	inter, found := currentStates.Get(id)

	if !found {
		fmt.Fprintln(w, "PDA Does Not Exists!")
	} else {

		currentState := inter.(string)
		fmt.Fprintln(w, currentState)
	}

}

func QueuedTokens(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	inter, found := tokenQueues.Get(id)

	if !found {
		fmt.Fprintf(w, "PDA Does Not Exists!")
	} else {

		queue := inter.(map[int]string)
		fmt.Fprint(w, queue)
	}

}

func SnapShot(w http.ResponseWriter, r *http.Request) {

	url := r.URL.Path
	id := strings.Split(url, "/")[3]
	kString := strings.Split(url, "/")[5]
	var k int
	if n, err := strconv.Atoi(kString); err == nil {
		k = n
	} else {
		fmt.Fprintln(w, kString, "is not an integer.")
	}

	inter, found := tokenQueues.Get(id)

	if !found {
		fmt.Fprintf(w, "PDA Does Not Exists!")
	} else {

		queue := inter.(map[int]string)
		fmt.Fprintln(w, queue)

		interStack, _ := stacks.Get(id)
		stack := interStack.(Stack)
		output := peek(k, stack)
		fmt.Fprintln(w, output)

		inter, _ := currentStates.Get(id)
		currentState := inter.(string)
		fmt.Fprintln(w, currentState)
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
		fmt.Fprintf(w, "PDA Does Not Exists!")

	} else {

		currentStates.Delete(id)
		pdas.Delete(id)
		tokenQueues.Delete(id)
		counts.Delete(id)
		stacks.Delete(id)

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
	http.ListenAndServe(":8080", r)
}
