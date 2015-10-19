package util

import (
  "bytes"
  "errors"
  "fmt"
  "io/ioutil"
  "log"
  "os"
  "regexp"
  "strings"
)

import (
  "golang.org/x/net/html"
)

import (
  "github.com/jamoozy/util/lg"
)


////////////////////////////////////////////////////////////////////////////////
//                        Special Logger (debug only)                         //
////////////////////////////////////////////////////////////////////////////////

// Logger responsible for logging traversal into and out of an HTML tree.
var nodeLogger struct {
  *log.Logger
  indent string
}

// General logger.
var l *lg.Logger

// Sets up the util -- may panic if there's an error opening up "node.log" for
// writing logs.
func init() {
  file, err := os.Create("node.log")
  if err != nil {
    panic(err.Error())
  }
  nodeLogger.Logger = log.New(file, "", 0)
  l = lg.New(os.Stdout, "[utl] ", 0, 0, false)
}

// Logs having visited a node.
func logNode(n *html.Node) {
  switch n.Type {
  case html.ElementNode:
    str := nodeLogger.indent + str(n)
    nodeLogger.Logger.Println(str)
    nodeLogger.indent += "  "
  default:
    nodeLogger.Logger.Printf(`%sSkipping "%s"`, nodeLogger.indent, str(n))
  }
}

// Logs having left a node.
func logExit(n *html.Node) {
  switch n.Type {
  case html.ElementNode:
    nodeLogger.indent = nodeLogger.indent[2:]
    str := fmt.Sprintf("%s</%s>", nodeLogger.indent, n.Data)
    nodeLogger.Logger.Println(str)
  }
}



////////////////////////////////////////////////////////////////////////////////
//                             Prettifying Nodes                              //
////////////////////////////////////////////////////////////////////////////////

func str(n *html.Node) string {
  if n == nil {
    return "nil"
  }
  switch n.Type {
  case html.ErrorNode:
    return fmt.Sprintf(`ErrNode("%s")`, n.Data)
  case html.TextNode:
    return fmt.Sprintf("%dB of text", len([]byte(n.Data)))
  case html.DocumentNode:
    return fmt.Sprintf("Document")
  case html.ElementNode:
    strs := make([]string, 0, len(n.Attr) + 1)
    strs = append(strs, n.Data)
    for _, attr := range n.Attr {
      strs = append(strs, fmt.Sprintf(`%s="%s"`, attr.Key, attr.Val))
    }
    return "<" + strings.Join(strs, " ") + ">"
  case html.CommentNode:
    return fmt.Sprintf("%dB of comments", len([]byte(n.Data)))
  case html.DoctypeNode:
    return fmt.Sprintf("doctype:%s", n.Data)
  }
  panic(`Invalid node type.`)
}



////////////////////////////////////////////////////////////////////////////////
//                                Main Parser                                 //
////////////////////////////////////////////////////////////////////////////////

// ---- Convenience.

// Determines if the file with the given file name, `fname`, exists.
func Exists(fname string) bool {
  if _, err := os.Stat(fname); os.IsNotExist(err) {
    return false
  }
  return true
}

// Gets the attribute with the given key from `node`.  Returns the empty string
// if not found, or if `node` is `nil`.
//
// TODO test me
func GetAttr(node *html.Node, attrKey string) string {
  for _, attr := range node.Attr {
    if attrKey == attr.Key {
      return attr.Val
    }
  }

  return ""
}

// Gets the content of the node as a text string.
//
// TODO test me
func GetText(node *html.Node) string {
  buf := bytes.NewBufferString("")
  for n := node.FirstChild; n != nil; n = n.NextSibling {
    if n.Type == html.TextNode {
      buf.WriteString(n.Data)
    } else {
      s := str(n)
      buf.WriteString(fmt.Sprintf("\n%s\n", s))
      buf.WriteString(GetText(n))
      buf.WriteString(fmt.Sprintf("\n<~%s\n", s))
    }
  }

  s, err := ioutil.ReadAll(buf)
  if err != nil {
    fmt.Printf(err.Error())
    return ""
  }

  return string(s)
}

// Determines if the node is a match for the described node.
func isMatch(n *html.Node, name, id, class string) (res bool) {
  if !(n.Type == html.ElementNode && n.Data == name) {
    return false
  }

  attrs := make(map[string]string)
  for _, attr := range n.Attr {
    attrs[attr.Key] = attr.Val
  }

  l.Printf("Got %s", attrs)

  if id != "" {
    l.Printf(`id = "%s"`, id)
    if id != attrs["id"] {
      l.Printf(`attrs["id"] = %s`, attrs["id"])
      return false
    }
  }

  if class != "" {
    l.Printf(`class = "%s"`, class)
    if class != attrs["class"] {
      l.Printf(`attrs["class"] = %s`, attrs["class"])
      return false
    }
  }

  return true
}


// ---- Find recursive algorithm.

// Finds the set of nodes that conform to the passed CSS selector, `path`.
func Find(root *html.Node, path string) (nodes []*html.Node, err error) {
  return find(root, map[string]bool{path: true})
}

// Finds the set of nodes that conform to the passed CSS selectors, `paths`.
func find(root *html.Node, paths map[string]bool) (nodes []*html.Node, err error) {
  // Rebuild the HTML document (debugging/verification step).
  logNode(root)
  defer logExit(root)

  if root == nil {
    msg := fmt.Sprintf(`find(nil, "%v"): err`, paths)
    lg.EnterExit(msg)
    return nil, errors.New(msg)
  }
  lg.Enter(`find(%s, "%v")`, str(root), paths)

  // Regular expression used to parse CSS.
  css := regexp.MustCompile(`(\w+)(#(\w+))?(\.(\w+))?(\s+(.*))?`)

  nextPaths := map[string]bool{}
  for k, v := range paths {
    nextPaths[k] = v
  }

  // Ensures this node is only added once (see end of for loop).
  addNode := false

  // Determine if this node matches any of these paths, and add sub-paths as
  // warranted.
  for path, _ := range paths {
    // Parse out the next name, id, and class from the CSS.
    md := css.FindStringSubmatch(path)
    if md == nil {
      lg.Exit("Invalid path.")
      return nil, errors.New("Invalid path.")
    }
    name, id, class, remain := md[1], md[3], md[5], md[7]
    l.Printf(`matched: "%s", "%s", "%s", "%s"`, name, id, class, remain)

    // Build next paths.
    if isMatch(root, name, id, class) {
      // If it's a match, `remain` becomes a valid path to search for in
      // subnodes.
      if remain == "" {
        addNode = true
        l.Printf(`%s fits "%v"`, str(root), paths)
      } else {
        nextPaths[remain] = true
      }
    }
    l.Printf("nextPaths: %v", nextPaths)
  }

  if addNode {
    nodes = append(nodes, root)
  }

  // The recursive bit.  Check all sub-children for both the same pattern and
  // the previous patterns.
  for c := root.FirstChild; c != nil; c = c.NextSibling {
    subNodes, err := find(c, nextPaths)
    if err != nil {
      lg.Exit("passing along err: " + err.Error())
      return nil, err
    }
    if subNodes != nil {
      l.Printf("Appending %d nodes.", len(subNodes))
      nodes = append(nodes, subNodes...)
    }
  }

  lg.Exit(`find(%s, %v): %d nodes`, str(root), paths, len(nodes))
  return nodes, nil
}
