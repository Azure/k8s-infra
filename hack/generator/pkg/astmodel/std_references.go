package astmodel

var (
	GopterReference     PackageReference = MakeExternalPackageReference("github.com/leanovate/gopter")
	GopterGenReference  PackageReference = MakeExternalPackageReference("github.com/leanovate/gopter/gen")
	GopterPropReference PackageReference = MakeExternalPackageReference("github.com/leanovate/gopter/prop")
	GomegaReference     PackageReference = MakeExternalPackageReference("github.com/onsi/gomega")
	JsonReference       PackageReference = MakeExternalPackageReference("encoding/json")
	PrettyReference     PackageReference = MakeExternalPackageReference("github.com/kr/pretty")
	ReflectReference    PackageReference = MakeExternalPackageReference("reflect")
	TestingReference    PackageReference = MakeExternalPackageReference("testing")

	GomegaImport PackageImport = NewPackageImport(GomegaReference).WithName(".")
)
