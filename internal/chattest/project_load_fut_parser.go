package chattest

import (
	"go/ast"
	"go/printer"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

type PackageID = string

type focalMethodParser struct {
	pkgsMap map[PackageID]*packages.Package
}

func (fmp *focalMethodParser) Parse(pkgs []*packages.Package, cfg *Config) (*FocalMethod, error) {
	fmp.pkgsMap = make(map[PackageID]*packages.Package)
	for _, pkg := range pkgs {
		fmp.pkgsMap[pkg.ID] = pkg
	}

	return fmp.findFocalMethod(cfg.FuncPath, cfg.FuncName)
}

// Todo: findFocalMethod should take into account the receiver name
func (fmp *focalMethodParser) findFocalMethod(funcPath, funcName string) (*FocalMethod, error) {
	fut := &FocalMethod{}

	for _, pkg := range fmp.pkgsMap {
		for _, file := range pkg.Syntax {
			pos := pkg.Fset.Position(file.Pos())
			if pos.Filename != funcPath {
				continue
			}
			for _, decl := range file.Decls {
				if decl, ok := decl.(*ast.FuncDecl); ok && decl.Name.Name == funcName {
					body := &strings.Builder{}
					printer.Fprint(body, pkg.Fset, decl)

					fut = &FocalMethod{
						Name: decl.Name.Name,
						Body: body.String(),
						Pkg: Pkg{
							ID:   pkg.ID,
							Name: pkg.Name,
						},
						File: pkg.Fset.File(file.Pos()).Name(),
						Uses: fmp.extractUses(pkg, decl),
					}
				}
			}
		}
	}

	return fut, nil
}

func (fmp *focalMethodParser) extractUses(pkg *packages.Package, decl *ast.FuncDecl) map[QualifiedName]Definition {
	uses := make(map[QualifiedName]Definition)
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		if node, ok := n.(*ast.Ident); ok {
			if obj := pkg.TypesInfo.ObjectOf(node); obj != nil && obj.Pkg() != nil {
				fmp.processObject(obj, node, uses)
			}
		}
		return true
	})
	return uses
}

func (fmp *focalMethodParser) processObject(obj types.Object, node *ast.Ident, uses map[QualifiedName]Definition) {
	switch objType := obj.Type().(type) {
	case *types.Signature:
		fmp.processSignature(obj, objType, node, uses)
	case *types.Named:
		fmp.processNamedType(obj, objType, uses)
	case *types.Pointer:
		if namedType, ok := objType.Elem().(*types.Named); ok {
			fmp.processNamedType(obj, namedType, uses)
		}
	}
}

func (fmp *focalMethodParser) processSignature(obj types.Object, objType *types.Signature, node *ast.Ident, uses map[QualifiedName]Definition) {
	if obj.Pkg() == nil {
		return
	}

	p, found := fmp.pkgsMap[obj.Pkg().Path()]
	if !found {
		return
	}

	recv := fmp.processReceiver(objType.Recv())
	var recvName *string
	if recv != nil {
		recvName = &recv.Name
	}

	var fnBody *string
	var fnFilePath *OsPath
	if recv == nil || !recv.IsInterface {
		typeDef := findFuncDefinition(p, recvName, node.Name)
		fnBody = &typeDef.body
		fnFilePath = &typeDef.filePath
	}

	funcDef := FuncDef{
		id:   QualifyFuncName(obj.Pkg().Path(), recvName, node.Name),
		name: node.Name,
		pkg: Pkg{
			ID:   obj.Pkg().Path(),
			Name: obj.Pkg().Name(),
		},
		typ:  obj.Type().String(),
		body: fnBody,
		file: fnFilePath,
		recv: recv,
	}
	uses[funcDef.id] = funcDef
}

func (fmp *focalMethodParser) processReceiver(recv *types.Var) *Recv {
	if recv == nil {
		return nil
	}
	if namedType, ok := recv.Type().(*types.Named); ok && namedType.Obj() != nil {
		_, isInterface := namedType.Underlying().(*types.Interface)
		return &Recv{
			ID:          QualifyTypeName(namedType.Obj().Pkg().Path(), namedType.Obj().Name()),
			Name:        namedType.Obj().Name(),
			IsInterface: isInterface,
			Pkg: Pkg{
				ID:   namedType.Obj().Pkg().Path(),
				Name: namedType.Obj().Pkg().Name(),
			},
		}
	}
	return nil
}

func (fmp *focalMethodParser) processNamedType(obj types.Object, objType *types.Named, uses map[QualifiedName]Definition) {
	if objType.Obj().Pkg() == nil {
		return
	}

	p, found := fmp.pkgsMap[objType.Obj().Pkg().Path()]
	if !found {
		return
	}
	typeDef := findTypeDefinition(p, objType.Obj().Name())

	def := TypeDef{
		id:   QualifyTypeName(obj.Pkg().Path(), obj.Name()),
		name: obj.Name(),
		pkg: Pkg{
			ID:   objType.Obj().Pkg().Path(),
			Name: objType.Obj().Pkg().Name(),
		},
		body: typeDef.body,
		file: typeDef.filePath,
	}
	uses[def.id] = def
}

type typeDefinition struct {
	name     string
	pkg      *packages.Package
	filePath OsPath
	body     string
}

func findTypeDefinition(pkg *packages.Package, declName string) *typeDefinition {
	return findDefinition(pkg, declName, nil)
}

func findFuncDefinition(pkg *packages.Package, receiver *string, fnName string) *typeDefinition {
	return findDefinition(pkg, fnName, receiver)
}

func findDefinition(pkg *packages.Package, declName string, receiver *string) *typeDefinition {
	buf := &strings.Builder{}
	var filePath OsPath

	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				if receiver == nil {
					if typeSpec, ok := findTypeSpec(decl, declName); ok {
						printer.Fprint(buf, pkg.Fset, typeSpec)
						filePath = OsPath(pkg.Fset.File(file.Pos()).Name())
					}
				}
			case *ast.FuncDecl:
				if isMatchingFuncDecl(decl, declName, receiver) {
					printer.Fprint(buf, pkg.Fset, decl)
					filePath = OsPath(pkg.Fset.File(file.Pos()).Name())
				}
			}
		}
	}

	return &typeDefinition{
		name:     declName,
		pkg:      pkg,
		filePath: filePath,
		body:     buf.String(),
	}
}

func findTypeSpec(decl *ast.GenDecl, declName string) (*ast.TypeSpec, bool) {
	for _, spec := range decl.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			if typeSpec.Name.Name == declName {
				return typeSpec, true
			}
		}
	}
	return nil, false
}

func isMatchingFuncDecl(decl *ast.FuncDecl, fnName string, receiver *string) bool {
	if decl.Name != nil && decl.Name.Name == fnName {
		if receiver != nil {
			if decl.Recv == nil || len(decl.Recv.List) == 0 || getReceiverName(decl) != *receiver {
				return false
			}
		}
		return true
	}
	return false
}

func getReceiverName(funcDecl *ast.FuncDecl) string {
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		recvType := funcDecl.Recv.List[0].Type

		switch expr := recvType.(type) {
		case *ast.Ident:
			return expr.Name
		case *ast.StarExpr:
			if ident, ok := expr.X.(*ast.Ident); ok {
				return ident.Name
			}
		}
	}
	return ""
}
