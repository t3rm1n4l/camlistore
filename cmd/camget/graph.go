/*
Copyright 2012 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"camlistore.org/pkg/blobref"
	"camlistore.org/pkg/client"
	"camlistore.org/pkg/index"
	"camlistore.org/pkg/schema"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type node struct {
	br *blobref.BlobRef
	g  *graph

	size  int64
	ss    *schema.Superset
	edges []*blobref.BlobRef
}

func (n *node) dotName() string {
	return strings.Replace(n.br.String(), "-", "_", -1)
}

func (n *node) dotLabel() string {
	name := n.displayName()
	if n.ss == nil {
		return fmt.Sprintf("%s\n%d bytes", name, n.size)
	}
	return name + "\n" + n.ss.Type
}

func (n *node) color() string {
	if n.br.Equal(n.g.root) {
		return "#a0ffa0"
	}
	if n.ss == nil {
		return "#aaaaaa"
	}
	return "#a0a0ff"
}

func (n *node) displayName() string {
	s := n.br.String()
	s = s[strings.Index(s, "-")+1:]
	return s[:7]
}

func (n *node) load() {
	defer n.g.wg.Done()
	rc, err := fetch(n.g.c, n.br)
	check(err)
	defer rc.Close()
	sniff := new(index.BlobSniffer)
	n.size, err = io.Copy(sniff, rc)
	check(err)
	sniff.Parse()
	ss, ok := sniff.Superset()
	if !ok {
		return
	}
	n.ss = ss
	for _, part := range ss.Parts {
		n.addEdge(part.BlobRef)
		n.addEdge(part.BytesRef)
	}
}

func (n *node) addEdge(dst *blobref.BlobRef) {
	if dst == nil {
		return
	}
	n.g.startLoadNode(dst)
	n.edges = append(n.edges, dst)
}

type graph struct {
	c    *client.Client
	root *blobref.BlobRef

	mu sync.Mutex // guards n
	n  map[string]*node

	wg sync.WaitGroup
}

func (g *graph) startLoadNode(br *blobref.BlobRef) {
	g.mu.Lock()
	defer g.mu.Unlock()
	key := br.String()
	if _, ok := g.n[key]; ok {
		return
	}
	n := &node{
		g:  g,
		br: br,
	}
	g.n[key] = n
	g.wg.Add(1)
	go n.load()
}

func printGraph(c *client.Client, root *blobref.BlobRef) {
	g := &graph{
		c:    c,
		root: root,
		n:    make(map[string]*node),
	}
	g.startLoadNode(root)
	g.wg.Wait()
	fmt.Println("digraph G {")
	fmt.Println(" node [fontsize=10,fontname=Arial]")
	fmt.Println(" edge [fontsize=10,fontname=Arial]")

	for _, n := range g.n {
		fmt.Printf("\n  %s [label=%q,style=filled,fillcolor=%q]\n", n.dotName(), n.dotLabel(), n.color())
		for i, e := range n.edges {
			// TODO: create an edge type.
			// Also, this edgeLabel is specific to file parts.  Other schema
			// types might not even have a concept of ordering.  This is hack.
			edgeLabel := fmt.Sprintf("%d", i)
			if i == 0 {
				edgeLabel = "first"
			} else if i == len(n.edges)-1 {
				edgeLabel = "last"
			}
			fmt.Printf("  %s -> %s [label=%q]\n", n.dotName(), g.n[e.String()].dotName(), edgeLabel)
		}
	}
	fmt.Printf("}\n")
}
