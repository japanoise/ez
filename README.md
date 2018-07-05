# ez

`ez` is an interpreter for an unstructured dialect of BASIC, written in
Golang. It is inspired by the BASIC interpreters you get when firing up e.g. a
Speccy or a C64.

## ...why

Why not? BASIC is fun and simple, and pretty easy to write an interpreter for.

## Usage

For an interactive session: `ez`

To load a program: `ez [file]`

- Keywords are case insensitive
- Variables are case sensitive
- In general, everything needs to be space-separated.
- Integer and string variables are supported - use a suffix of $ for a string
  variable, and no suffix for an integer variable.
- Comparing strings and integers uses the length of the string for comparison
- No looping constructs - these can be constructed with `IF [cond] THEN GOTO [line]`
- Setting a variable is done through the `LET` keyword. Multiple variables can
  be set inside a `LET` at once by using the separator `;`
- The `END` keyword is not mandatory, but it's useful.

## Examples

Hello World:

```
10 PRINT "Hello world!"
20 END
```

The BASIC shout-out from Futurama S01E03:

```
10 PRINT "HOME"
20 PRINT "SWEET"
30 GOTO 10
```

The slightly more complicated Hello User:

```
10 INPUT "What is your name?" u$
20 IF u$ = 0 GOTO 10
30 PRINT "Hello " u$
40 END
```

Rough port of [a program from Wikipedia][wiki]

```
10 INPUT "What is your name? " u$
20 PRINT "Hello " u$
30 INPUT "How many stars do you want? " n
40 LET s$ = ""
50 LET i = 1
60 LET s$ = s$ + "*"
70 LET i = i + 1
80 IF i <= n THEN GOTO 60
90 PRINT s$
100 INPUT "Do you want more stars? " a$
110 IF a$ = 0 THEN GOTO 100
120 IF a$ = "Y" THEN GOTO 30
130 IF a$ = "y" THEN GOTO 30
140 PRINT "Goodbye " u$
150 END
```

## Copying

`ez` is licensed MIT, as are the BASIC snippets in this README.

[wiki]: https://en.wikipedia.org/wiki/BASIC#Unstructured_BASIC
