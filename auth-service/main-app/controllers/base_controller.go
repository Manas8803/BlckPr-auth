package controllers

import "github.com/gin-gonic/gin"

func BaseRoute(c *gin.Context) {
	html := `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Auth API</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					background-color: #e6f0ff; /* Light blue background */
					text-align: center;
					margin: 0;
					display: flex;
					justify-content: center;
					align-items: center;
					height: 100vh;
				}

				.container {
					background-color: #fff;
					padding: 20px;
					border-radius: 10px;
					box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
				}

				h1 {
					color: #333;
				}

				p {
					color: #555;
				}

				a {
					display: inline-block;
					margin-top: 20px;
					padding: 10px 20px;
					background-color: #007bff; /* Blue color for link */
					color: #fff;
					text-decoration: none;
					border-radius: 5px;
					transition: background-color 0.3s;
				}

				a:hover {
					background-color: #0056b3; /* Darker blue on hover */
				}
			</style>
		</head>
		<body >
			<div class="container" style="width:20rem;">
				<h3 style="font-family:monospace;font-size:2rem;">Auth API</h3>
				<a href="api/v1/docs/index.html#/example" style="font-family:Arial;font-size:1rem;letter-spacing:2px;">Documentation</a>
			</div>
		</body>
		</html>
		`
	c.Data(200, "text/html; charset=utf-8", []byte(html))
}
