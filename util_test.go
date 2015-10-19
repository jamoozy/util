package util


import (
  "bytes"
  "fmt"
  "golang.org/x/net/html"
  "io/ioutil"
  "testing"
)

func assert(t *testing.T, cond bool, msg string, args ...interface{}) {
  if !cond {
    t.Log(fmt.Sprintf(msg, args...))
    t.FailNow()
  }
}

func getRootHtml(t *testing.T) *html.Node {
  contents, err := ioutil.ReadFile("page.html")
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

func runTest(t *testing.T, path string, num int) {
  nodes, err := Find(getRootHtml(t), path)
  if err != nil {
    t.Error(err)
    t.FailNow()
  }
  if nodes == nil {
    t.Log(`nodes was nil`)
    t.Fatal()
  }

  assert(t, len(nodes) == num, "len(nodes):%d exp:%d", len(nodes), num)
}

func TestFindD  (t *testing.T) { runTest(t, "div", 11) }
func TestFindDa (t *testing.T) { runTest(t, "div a", 15) }
func TestFindDi (t *testing.T) { runTest(t, "div img", 13) }
func TestFindDai(t *testing.T) { runTest(t, "div a img", 10) }
