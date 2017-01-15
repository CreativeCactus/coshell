package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestMain(t *testing.T) {

	testCompare(t,"Test basic operation",
		testHelper([]string{"coshell"}, "echo a\necho b\necho c\n"),
		"a\nb\nc\n",
	)

	testCompare(t,"Test the prefix string",
		testHelper([]string{"coshell", "-p", "echo "}, "a\nb\nc\n"),
		"a\nb\nc\n",
	)

	testCompare(t,"Test the default escape string",
		testHelper([]string{"coshell", "-p", "echo ?{} ?{3#3}", "-e"}, "a\nb\nc\n"),
		"a 003\nb 004\nc 005\n",
	)

}

func testCompare(t *testing.T, description, result, expect string){
	if result != expect {
		t.Fatalf("%s:\n\t%q got, expected:\n\t%q\n",description,result,expect)
	} else {
		t.Logf("[PASSED] %s",description)
	}
}

func testHelper(args []string, input string) string {
	os.Args = args
	ro, wo, _ := os.Pipe()
	ri, wi, _ := os.Pipe()
	oldStdout := os.Stdout
	oldStdin := os.Stdin // connected to /dev/null in tests, anyway
	os.Stdout = wo
	os.Stdin = ri
	defer func() {
		os.Stdout = oldStdout
		os.Stdin = oldStdin
	}()
	outC := make(chan string)
	done := make(chan int)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		//var buf bytes.Buffer
		//io.Copy(&buf, ro)
		buf, err := ioutil.ReadAll(ro)
		if err!=nil {
			panic(err)
		}
		outC <- string(buf)
	}()
	go func (){
		coshell()
		wo.Close()
		done <- 1
	}()
	wi.Write([]byte(input))
	wi.Close()
	<- done
	out := <-outC
	return out
}
