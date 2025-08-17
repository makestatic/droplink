package cli

type Commands struct {
	Port     int    `short:"p" name:"port" default:"8080" desc:"Port to listen on"`
	Global   bool   `short:"g" name:"global" default:"false" desc:"Generate public link"`
	Zip      bool   `short:"z" name:"zip" default:"false" desc:"Zip the given input before uploading"`
	Password string `short:"P" name:"password" default:"" desc:"Encrypt the given input by password"`
	Timeout  int    `short:"t" name:"timeout" default:"0" desc:"Timeout in seconds, 0 for no timeout"`
	// Input    string/*os.File*/ `short:"i" name:"input" default:"." desc:"Input files or directories to upload"`
}
