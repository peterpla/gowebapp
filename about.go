package main

import (
	"fmt"
	"net/http"
)

func (srv *server) handleAbout(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Content-Security-Policy", "script-src https://stackpath.bootstrapcdn.com https://ajax.googleapis.com https://cdnjs.cloudflare.com; object-src 'none'")

	fmt.Fprintf(w, `<!DOCTYPE html>
	<html lang="en">
	<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	<title>Bootstrap 4 Responsive Layout Example</title>
	<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css">
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css">
	<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.4.1/jquery.min.js"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js"></script>
	<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js"></script>
	</head>
	<body>
	<nav class="navbar navbar-expand-md navbar-dark bg-dark mb-3">
		<div class="container-fluid">
			<a href="#" class="navbar-brand mr-3">Tutorial Republic</a>
			<button type="button" class="navbar-toggler" data-toggle="collapse" data-target="#navbarCollapse">
				<span class="navbar-toggler-icon"></span>
			</button>
			<div class="collapse navbar-collapse" id="navbarCollapse">
				<div class="navbar-nav">
					<a href="home" class="nav-item nav-link active">Home</a>
					<a href="#" class="nav-item nav-link">Services</a>
					<a href="#" class="nav-item nav-link">AboutXXX</a>
					<a href="#" class="nav-item nav-link">Contact</a>
				</div>
				<div class="navbar-nav ml-auto">
					<a href="#" class="nav-item nav-link">Register</a>
					<a href="#" class="nav-item nav-link">Login</a>
				</div>
			</div>
		</div>    
	</nav>
	<div class="container">
		<div class="jumbotron">
			<h1>About</h1>
			<p class="lead">Lorem ipsum dolor sit amet, consectetur adipiscing elit. Ut suscipit vel felis in interdum. Suspendisse in quam lobortis, ultrices lorem vel, scelerisque nibh. Donec lobortis, odio sit amet cursus blandit, purus ex blandit arcu, vel euismod libero ligula quis sem. Mauris in cursus nunc, ut laoreet lorem. Sed id tempor ante.</p>
		</div>
		<hr>
		<footer>
			<div class="row">
				<div class="col-md-6">
					<p>Copyright &copy; 2019 Tutorial Republic</p>
				</div>
				<div class="col-md-6 text-md-right">
					<a href="#" class="text-dark">Terms of Use</a> 
					<span class="text-muted mx-2">|</span> 
					<a href="#" class="text-dark">Privacy Policy</a>
				</div>
			</div>
		</footer>
	</div>
	</body>
	</html>`)
}
