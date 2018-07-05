package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

// InputString ...
// Prompts for a string
func InputString(prompt string) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	for scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

// InputNumber ...
// Prompts for an integer
func InputNumber(prompt string) (int, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	for scanner.Scan() {
		num, err := strconv.Atoi(scanner.Text())
		if err == nil {
			return num, nil
		}
		fmt.Print(prompt)
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, nil
}
