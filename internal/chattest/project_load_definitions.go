package chattest

import "strings"

// QualifiedName is an identifier for golang objects.
// It prefixes object's name with the package path.
// For functions it also includes the receiver name.
type QualifiedName string

func QualifyTypeName(pkgPath string, objName string) QualifiedName {
	return QualifiedName(pkgPath + "." + objName)
}

// Todo: this is not correct, as it doesn't take into account if the receiver
// is a pointer or not.
// i.e. it doesn't make a difference between func (t T) and func (t *T)
func QualifyFuncName(pkgPath string, receiver *string, objName string) QualifiedName {
	var sb strings.Builder
	sb.WriteString(pkgPath)
	if receiver != nil {
		sb.WriteString(".")
		sb.WriteString(*receiver)
	}
	sb.WriteString(".")
	sb.WriteString(objName)
	return QualifiedName(sb.String())
}

type OsPath = string

// Definition encapsulates some info about some golang objects.
// For the purpose of this app, it can hold either a function or a type definition.
type Definition interface {
	ID() QualifiedName
	Name() string
	Package() Pkg
	// Not all definitions have a body, i.e. a func of an interface
	Body() *string
	File() *OsPath
}

type Recv struct {
	ID          QualifiedName
	Name        string
	Pkg         Pkg
	IsInterface bool
}

type FuncDef struct {
	id   QualifiedName
	name string
	pkg  Pkg
	typ  string
	body *string
	file *OsPath
	recv *Recv
}

func (f FuncDef) ID() QualifiedName {
	return f.id
}

func (f FuncDef) Name() string {
	return f.name
}

func (f FuncDef) Package() Pkg {
	return f.pkg
}

func (f FuncDef) Body() *string {
	return f.body
}

func (f FuncDef) File() *OsPath {
	return f.file
}

type TypeDef struct {
	id   QualifiedName
	name string
	pkg  Pkg
	body string
	file OsPath
}

func (s TypeDef) ID() QualifiedName {
	return s.id
}

func (s TypeDef) Name() string {
	return s.name
}

func (s TypeDef) Package() Pkg {
	return s.pkg
}

func (s TypeDef) Body() *string {
	return &s.body
}

func (s TypeDef) File() *OsPath {
	return &s.file
}
