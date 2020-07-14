# powerball-numbers-generator-and-api

# Disclaimer
I started creating this project with zero knowledge of Go. I'm refactoring it along with my growing knowledge and understanding of Go and its philosophy.

## OverviewS
Powerball combination generator and web UI to display the data. 

Backend is written in Go and has few main components. 

- Web server that serves API.
- Web server that serves React js.
- Random combination generator that generates combinations and sends it to goroutine that inserts it in a DB and web socket broadcaster.
- Web socket handler, that broadcasts randomly generated combination to web clients.
- Powerball parser that parses recent real powerball drawings and saves it to the DB.

Project also contains react application source and generator_txt, which is standalone combination generator. 

I don't have much time to dedicate it to the project, but I in love with Go and enjoy creating things using it.
This project might not have much sense, but I was happy creating it and learning Go.

## generator_txt
### generator.go
Generates powerball numbers and saves them in the file when you stop it. To stop it, create a file named stop in the directory with the script. 

I've run it on 8 core, 32 gb ram  machine, and it didn't crush in a week, then I stop it. It consumes about 30gb of ram memory. Definitely, there is room for improvement. Dump file is about 6.2GB with all possible combinations. 
### inserter.go
Inserts data from dump file into mysql DB. `schema.sql` could be found in project's root directory.

## web
### api.go
Entry point - declares API logic and applications initialization. (*going to be refactored and application init will be moved into separate file.)
### db.go
DB initialization logic, combinationsData strut definition, reset counts definition
### generator.go
Contains logic for generating combinations, saving combinations into the DB, and broadcasting it to the websocket's clients.
### react.go
Handler for serving react application.
### parser.go
Contains logic for parsing and saving information into the DB
### websockets.go
Logic and handler for websockets. 
### logger.go
Logger for web server requests.S