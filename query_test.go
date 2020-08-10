package htmlquery

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/antchfx/xpath"
	"golang.org/x/net/html"
)

const htmlSample = `<!DOCTYPE html><html lang="en-US">
<head>
<title>Hello,World!</title>
</head>
<body>
<div class="container">
<header>
	<!-- Logo -->
   <h1>City Gallery</h1>
</header>  
<nav>
  <ul>
    <li><a href="/London">London</a></li>
    <li><a href="/Paris">Paris</a></li>
    <li><a href="/Tokyo">Tokyo</a></li>
  </ul>
</nav>
<article>
  <h1>London</h1>
  <img src="pic_mountain.jpg" alt="Mountain View" style="width:304px;height:228px;">
  <p>London is the capital city of England. It is the most populous city in the  United Kingdom, with a metropolitan area of over 13 million inhabitants.</p>
  <p>Standing on the River Thames, London has been a major settlement for two millennia, its history going back to its founding by the Romans, who named it Londinium.</p>
</article>
<footer>Copyright &copy; W3Schools.com</footer>
</div>
</body>
</html>
`

var testDoc = loadHTML(htmlSample)

func BenchmarkSelectorCache(b *testing.B) {
	DisableSelectorCache = false
	for i := 0; i < b.N; i++ {
		getQuery("/AAA/BBB/DDD/CCC/EEE/ancestor::*")
	}
}

func BenchmarkDisableSelectorCache(b *testing.B) {
	DisableSelectorCache = true
	for i := 0; i < b.N; i++ {
		getQuery("/AAA/BBB/DDD/CCC/EEE/ancestor::*")
	}
}

func TestSelectorCache(t *testing.T) {
	SelectorCacheMaxEntries = 2
	for i := 1; i <= 3; i++ {
		getQuery(fmt.Sprintf("//a[position()=%d]", i))
	}
	getQuery("//a[position()=3]")

}

func TestLoadURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, htmlSample)
	}))
	defer ts.Close()

	_, err := LoadURL(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadDoc(t *testing.T) {
	tempHTMLdoc, err := ioutil.TempFile("", "sample_*.html")
	if err != nil {
		t.Fatal(err)
	}
	tempHTMLFilename := tempHTMLdoc.Name()
	defer func(tempHTMLdoc *os.File, filename string) {
		tempHTMLdoc.Close()
		os.Remove(filename)
	}(tempHTMLdoc, tempHTMLFilename)

	tempHTMLdoc.Write([]byte(htmlSample))

	if _, err := LoadDoc(tempHTMLFilename); err != nil {
		t.Fatal(err)
	}
}

func TestNavigator(t *testing.T) {
	testNewDoc := (*Node)(testDoc)
	top := testNewDoc.FindOne("//html")
	nav := top.CreateXPathNavigator()
	nav.MoveToChild() // HEAD
	nav.MoveToNext()
	if nav.NodeType() != xpath.TextNode {
		t.Fatalf("expectd node type is TextNode,but got %vs", nav.NodeType())
	}
	nav.MoveToNext() // <BODY>
	if nav.Value() != testNewDoc.FindOne("//body").InnerText() {
		t.Fatal("body not equal")
	}
	nav.MoveToPrevious() //
	nav.MoveToParent()   //<HTML>
	if (*Node)(nav.curr) != top {
		t.Fatal("current node is not html node")
	}
	nav.MoveToNextAttribute()
	if nav.LocalName() != "lang" {
		t.Fatal("node not move to lang attribute")
	}

	nav.MoveToParent()
	nav.MoveToFirst() // <!DOCTYPE html>
	if nav.curr.Type != html.DoctypeNode {
		t.Fatalf("expected node type is DoctypeNode,but got %d", nav.curr.Type)
	}
}

func TestXPath(t *testing.T) {
	testNewDoc := (*Node)(testDoc)
	node := testNewDoc.FindOne("//html")
	if av, err := node.AttributeValue("lang"); err != nil && av != "en-US" {
		t.Fatal("//html[@lang] != en-Us")
	}

	node = testNewDoc.FindOne("//header")
	if strings.Index(node.InnerText(), "Logo") > 0 {
		t.Fatal("InnerText() have comment node text")
	}
	if strings.Index(node.OutputHTML(true), "Logo") == -1 {
		t.Fatal("OutputHTML() shoud have comment node text")
	}
	link := testNewDoc.FindOne("//a[1]/@href")
	if link == nil {
		t.Fatal("link is nil")
	}
	if v := link.InnerText(); v != "/London" {
		t.Fatalf("expect value is /London, but got %s", v)
	}

}

func TestXPathCdUp(t *testing.T) {
	doc := loadHTML(`<html><b attr="1"></b></html>`)
	node := doc.FindOne("//b/@attr/..")
	t.Logf("node = %#v", node)
	if node == nil || node.Data != "b" {
		t.Fatal("//b/@id/.. != <b></b>")
	}
}

func TestQueryAll(t *testing.T) {
	doc := loadHTML(`<html><b attr="1"></b><b attr="2"></b></html>`)
	nodes, err := doc.QueryAll("//b/@attr/..")
	if err != nil {
		t.Error(err)
	}
	t.Logf("node = %#v", nodes)
	for _, n := range nodes {
		if n.Data != "b" {
			t.Error(n.Data)
		}
	}

}

func TestQueryAllText(t *testing.T) {
	doc := loadHTML(`<html><b attr="1">你好</b><b attr="2">shit</b></html>`)
	nodes, err := doc.QueryAll("//b")
	if err != nil {
		panic(err)
	}
	for _, node := range nodes {
		if len(node.Text()) < 2 {
			t.Error(node.Text())
		}
	}
}

func loadHTML(str string) *Node {
	node, err := Parse(strings.NewReader(str))
	if err != nil {
		panic(err)
	}
	return node
}

func TestPointer(t *testing.T) {
	// doc := loadHTML(`<html><body><b attr="1">你好</b><b attr="2">shit</b></body></html>`)
	// nodes1, _ := doc.QueryAll("//b")
	// nodes2, _ := doc.QueryAll("//body//b")
	// t.Error(nodes1, nodes2)
	// t.Error(reflect.ValueOf(nodes1[1]).Pointer())
	// t.Error(reflect.ValueOf(nodes2[0]).Pointer())
}

func TestComment(t *testing.T) {

	// doctxt := `<!DOCTYPE html>
	// <html>
	// 	<head>
	// 		<meta charset="UTF-8">
	// 		<title>Comment类型</title>
	// 	</head>
	// 	<body>
	// 		 <div id="myDiv"><!--A comment --></div>
	// 	</body>
	// 	<script>
	// 		 /*
	// 		  注释在 DOM 中是通过 Comment 类型来表示的。Comment 节点具有下列特征：
	// 	nodeType 的值为 8；
	// 		nodeName 的值为"#comment"；
	// 		nodeValue 的值是注释的内容；
	// 	 parentNode 可能是 Document 或 Element；
	// 		不支持（没有）子节点。
	// 		Comment 类型与 Text 类型继承自相同的基类，因此它拥有除 splitText()之外的所有字符串操
	// 		作方法。与 Text 类型相似，也可以通过 nodeValue 或 data 属性来取得注释的内容。
	// 		  * */
	// 		var div = document.getElementById("myDiv");
	// 		var comment = div.firstChild;
	// 		console.log(comment.data);
	// 	</script>`
	// doctxt = "<td></td>"
	// doc := loadHTML(doctxt)
	// // d, _ := doc.Query("//div[@id='myDiv']")
	// // commtent := d.First()
	// t.Error(doc)

}
