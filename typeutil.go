package selectivetesting

import "go/types"

func getUsedTypeNames(t types.Type) []*types.TypeName {
	typeNames := make([]*types.TypeName, 0)
	switch ct := t.(type) {
	case *types.Array:
		typeNames = append(typeNames, getUsedTypeNames(ct.Elem())...)
	case *types.Basic:
		break
	case *types.Chan:
		typeNames = append(typeNames, getUsedTypeNames(ct.Elem())...)
	case *types.Interface:
		for i := 0; i < ct.NumExplicitMethods(); i++ {
			typeNames = append(typeNames, getUsedTypeNames(ct.ExplicitMethod(i).Type())...)
		}
		for i := 0; i < ct.NumEmbeddeds(); i++ {
			typeNames = append(typeNames, getUsedTypeNames(ct.EmbeddedType(i))...)
		}
	case *types.Map:
		typeNames = append(typeNames, getUsedTypeNames(ct.Key())...)
		typeNames = append(typeNames, getUsedTypeNames(ct.Elem())...)
	case *types.Named:
		if obj := ct.Obj(); obj != nil {
			typeNames = append(typeNames, ct.Obj())
		}
	case *types.Pointer:
		typeNames = append(typeNames, getUsedTypeNames(ct.Elem())...)
	case *types.Signature:
		typeNames = append(typeNames, getUsedTypeNames(ct.Params())...)
		typeNames = append(typeNames, getUsedTypeNames(ct.Results())...)
	case *types.Slice:
		typeNames = append(typeNames, getUsedTypeNames(ct.Elem())...)
	case *types.Struct:
		for i := 0; i < ct.NumFields(); i++ {
			typeNames = append(typeNames, getUsedTypeNames(ct.Field(i).Type())...)
		}
	case *types.Tuple:
		for i := 0; i < ct.Len(); i++ {
			typeNames = append(typeNames, getUsedTypeNames(ct.At(i).Type())...)
		}
	case *types.TypeParam:
		if obj := ct.Obj(); obj != nil {
			typeNames = append(typeNames, ct.Obj())
		}
	case *types.Union:
		for i := 0; i < ct.Len(); i++ {
			typeNames = append(typeNames, getUsedTypeNames(ct.Term(i).Type())...)
		}
	}
	return typeNames
}
