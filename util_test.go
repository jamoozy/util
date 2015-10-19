package util


import (
  "bytes"
  "fmt"
  "golang.org/x/net/html"
  "io/ioutil"
  "testing"
)



////////////////////////////////////////////////////////////////////////////////
//                                    Find                                    //
////////////////////////////////////////////////////////////////////////////////

// If `cond` is false, fails the current test and prints out the formatted
// message.
func assert(t *testing.T, cond bool, msg string, args ...interface{}) {
  if !cond {
    t.Log(fmt.Sprintf(msg, args...))
    t.FailNow()
  }
}

// Gets the HTML node at the root of the tree in the default file.
func getRootHtml(t *testing.T) *html.Node {
  contents, err := ioutil.ReadFile("test_files/page.html")
  if err != nil {
    t.Error(err)
    t.FailNow()
  }

  root, err := html.Parse(bytes.NewReader(contents))
  if err != nil {
    t.Error(err)
    t.FailNow()
  }

  return root
}

// Runs a test for Find with `path` on the default page.  Expects `num` nodes as
// a result.
func runFindTest(t *testing.T, path string, num int) {
  nodes, err := Find(getRootHtml(t), path)
  if err != nil {
    t.Error(err)
    t.FailNow()
  }
  if nodes == nil {
    if num > 0 {
      t.Log(`nodes was nil`)
      t.Fatal()
    } else if num < 0 {
      t.Log(`expected negative number`)
      t.Fatal()
    }
  } else {
    assert(t, len(nodes) == num, "len(nodes):%d exp:%d", len(nodes), num)
  }
}

func TestFindScript (t *testing.T) { runFindTest(t, "foo", 0) }

// Find a "<div>".
func TestFindD  (t *testing.T) { runFindTest(t, "div", 11) }

// Find a "<div> ... <a>"
func TestFindDa (t *testing.T) { runFindTest(t, "div a", 15) }

// Find a "<div> ... <img>"
func TestFindDi (t *testing.T) { runFindTest(t, "div img", 13) }

// Find a "<div> ... <a> ... <img>"
func TestFindDai(t *testing.T) { runFindTest(t, "div a img", 10) }


