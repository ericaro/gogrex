package gogrex

import (
	"fmt"
	"testing"
)

var (
	exps = []string{
		"(a, b+ // titi \n )",
		"(a, b+ /* toto */ )",
		"(timing, ( id, value)+ )*, startDefinition,  (id, name )* , endDefinition",
		"(a,b)*",
		"x,(a,b)*",
		"(a,(b,c)*)*",
		"(a, b+ )*, end", // ko
		"(a, b*,c )",
		"(a, (b1,b2)*,c )",
		"(a, (b1,b2)*,c )+",
		"(a, b+ )",
		"(timing, (id,value)+)*", // it's ok
		"(timing, (id,value)+ )*, startDefinition, (id,name)*, endDefinition",
		"conf, (id, point)*, (timing, (id, temperature)* )+, endfile",
	}
	typs = []string{
		"itemError",
		"itemEOF",
		"itemStar",
		"itemPlus",
		"itemOpt",
		"itemSel",
		"itemSeq",
		"itemLeft",
		"itemRight",
		"itemIdentifier",
	}
	goldens = [][]item{
		[]item{ // "(a, b+ // titi \n )",
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "a"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "b"},
			item{typ: itemPlus, val: "+"},
			item{typ: itemComment, val: "// titi \n"},
			item{typ: itemRight, val: ")"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // "(a, b+ /* toto */ )",
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "a"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "b"},
			item{typ: itemPlus, val: "+"},
			item{typ: itemComment, val: "/* toto */"},
			item{typ: itemRight, val: ")"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // (timing, ( id, value)+ )*, startDefinition,  (id, name )* , endDefinition
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "timing"},
			item{typ: itemSeq, val: ","},
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "id"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "value"},
			item{typ: itemRight, val: ")"},
			item{typ: itemPlus, val: "+"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "startDefinition"},
			item{typ: itemSeq, val: ","},
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "id"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "name"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "endDefinition"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // (a,b)*
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "a"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "b"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // x,(a,b)*
			item{typ: itemIdentifier, val: "x"},
			item{typ: itemSeq, val: ","},
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "a"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "b"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // (a,(b,c)*)*
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "a"},
			item{typ: itemSeq, val: ","},
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "b"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "c"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // (a, b+ )*, end
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "a"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "b"},
			item{typ: itemPlus, val: "+"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "end"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // (a, b*,c )
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "a"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "b"},
			item{typ: itemStar, val: "*"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "c"},
			item{typ: itemRight, val: ")"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // (a, (b1,b2)*,c )
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "a"},
			item{typ: itemSeq, val: ","},
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "b1"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "b2"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "c"},
			item{typ: itemRight, val: ")"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // (a, (b1,b2)*,c )+
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "a"},
			item{typ: itemSeq, val: ","},
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "b1"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "b2"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "c"},
			item{typ: itemRight, val: ")"},
			item{typ: itemPlus, val: "+"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // (a, b+ )
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "a"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "b"},
			item{typ: itemPlus, val: "+"},
			item{typ: itemRight, val: ")"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // (timing, (id,value)+)*
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "timing"},
			item{typ: itemSeq, val: ","},
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "id"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "value"},
			item{typ: itemRight, val: ")"},
			item{typ: itemPlus, val: "+"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // (timing, (id,value)+ )*, startDefinition, (id,name)*, endDefinition
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "timing"},
			item{typ: itemSeq, val: ","},
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "id"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "value"},
			item{typ: itemRight, val: ")"},
			item{typ: itemPlus, val: "+"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "startDefinition"},
			item{typ: itemSeq, val: ","},
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "id"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "name"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "endDefinition"},
			item{typ: itemEOF, val: ""},
		},
		[]item{ // conf, (id, point)*, (timing, (id, temperature)* )+, endfile
			item{typ: itemIdentifier, val: "conf"},
			item{typ: itemSeq, val: ","},
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "id"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "point"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemSeq, val: ","},
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "timing"},
			item{typ: itemSeq, val: ","},
			item{typ: itemLeft, val: "("},
			item{typ: itemIdentifier, val: "id"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "temperature"},
			item{typ: itemRight, val: ")"},
			item{typ: itemStar, val: "*"},
			item{typ: itemRight, val: ")"},
			item{typ: itemPlus, val: "+"},
			item{typ: itemSeq, val: ","},
			item{typ: itemIdentifier, val: "endfile"},
			item{typ: itemEOF, val: ""},
		},
	}
)

func aTestLex(t *testing.T) {

	for j, s := range exps {
		fmt.Printf("[]item{ // %s\n", s)
		k := 0

		tokens := lex(s)

		for to := range tokens {
			// below to generate goldens
			//fmt.Printf("item{typ:%s,val:\"%s\"},\n", typs[i.typ], i.val)
			g := goldens[j][k]
			i := to.(item)
			if i.typ != g.typ || i.val != g.val {
				t.Fatal("unexpected Output %v %v %v", i, g.typ, g.val)
			}
			k++
			//typ, val := i.typ, i.val
		}
		fmt.Printf("},\n")

	}
}

//func TestShunting(t *testing.T) {
//	fmt.Printf("test shunting\n")
//	for _, s := range exps {
//		//fmt.Printf("[]item{ // %s\n", s)
//		fmt.Printf("shunting %s\n", s)
//		parse(s)
////		tokens := lex(s)
////		grammar, errchan := shunting(tokens)
////		for  {
////			select {
////			case t,ok := <-grammar:
////				fmt.Printf("    %s\n", t)
////				if !ok {
////					break
////				}
////			case err, ok := <-errchan:
////				fmt.Printf("  err  %v\n", err)
////				if !ok {
////					break
////				}
////			}
////		}
////		fmt.Printf("\n\n")
//
//	}
//}
