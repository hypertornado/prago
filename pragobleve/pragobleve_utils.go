package pragobleve

func must(err error) {
	if err != nil {
		panic(err)
	}
}
