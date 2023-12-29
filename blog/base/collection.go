package base

type RecipeBank struct {
	Posts       []Post
	UsedIDsPath string
}

type Collection map[string]RecipeBank
