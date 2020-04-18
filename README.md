# powerball-numbers-generator-and-api
## Overview
Powerball combination generator and some api written in golang as hobby project. 

This project was made late in a night and I didn't think much about variable or fucntion names, writing tests or anything else production like stuff. 
Probably, I will take care of that after I create a front end for it. 

Project idea came to me from nowhere and was a purpose to dive in Go <3

## generator_txt
### generator.go
Generates powerball numbers and saves them in the file when you stop it. To stop it create a file named stop in the directory with the script. 

I had it runing on 8 core machine with 32 gb machine and it haven't crushed in a week, then I stop it. It comsumes about 30gb of ram memory. Definately, there is room for improvement. Dump file is about 6GB. 
### inserter.go
Inserts results into mysql DB. `schema.sql` could be found in project's root directory.

## web
### api.go
Main file that containes api logic. 
### db.go
DB setup settings, same interface definition.
### combination_generator.go
generates and inserts powerball combinations into the DB. 
