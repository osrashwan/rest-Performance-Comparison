// Copyright (c) 2016, Gerasimos Maropoulos
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
//	  this list of conditions and the following disclaimer
//    in the documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse
//    or promote products derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL JULIEN SCHMIDT BE LIABLE FOR ANY
// DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
package iris

type tree struct {
	method     string
	rootBranch *Branch
	domain     string
	hosts      bool //if domain != "" we set it directly on .Plant
	cors       bool // if cross domain allow enabled
}

// Garden is the main area which routes are planted/placed
type Garden []tree

// getTreeByMethod returns the correct tree which it's method is equal to the given method, from the garden
func (g Garden) getTreeByMethod(method string) *tree {
	for _, _tree := range g {
		if _tree.method == method {
			return &_tree
		}
	}
	return nil
}

// getTreeByMethodAndDomain returns the correct tree which it's method&domain is equal to the given method&domain, from the garden
func (g Garden) getTreeByMethodAndDomain(method string, domain string) *tree {
	for _, _tree := range g {
		if _tree.domain == domain && _tree.method == method {
			return &_tree
		}
	}
	return nil
}

func (g Garden) getRootByMethod(method string) *Branch {
	for _, _tree := range g {
		if _tree.method == method {
			return _tree.rootBranch
		}
	}
	return nil
}

// getRootByMethodAndDomain returns the correct branch which it's method&domain is equal to the given method&domain, from a garden's tree
// trees with  no domain means that their domain==""
func (g Garden) getRootByMethodAndDomain(method string, domain string) *Branch {
	for _, _tree := range g {
		if _tree.domain == domain && _tree.method == method {
			return _tree.rootBranch
		}
	}
	return nil
}

// Plant plants/adds a route to the garden
func (g Garden) Plant(_route IRoute) Garden {

	//we have a domain to assign too
	theRoot := g.getRootByMethodAndDomain(_route.GetMethod(), _route.GetDomain())
	if theRoot == nil {
		theRoot = new(Branch)
		g = append(g, tree{_route.GetMethod(), theRoot, _route.GetDomain(), _route.GetDomain() != "", hasCors(_route)}) //hasCors is inside utils.go

	}
	theRoot.AddBranch(_route.GetDomain()+_route.GetPath(), _route.GetMiddleware())

	return g
}

// len returns the lengh of the trees inside this garden
func (g Garden) len() int {
	return len(g)
}

// get returns a tree from this garden by index
func (g Garden) get(index int) tree {
	return g[index]
}
