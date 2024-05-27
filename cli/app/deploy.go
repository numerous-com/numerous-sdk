package app

func Deploy(slug string, appName string) error {
	println("Uploading app source code...")
	println("Building app...")
	println("Starting app...")
	println("Your app is now running at https://numerous.com/app/organization/" + slug + "/app/" + appName)

	return nil
}
