## Step-by-Step Guide

### Step 1: Install Dependencies

First, you need to install the dependencies for your Go project. You can do this using the following command:

```bash
go mod download
```

This command will download all the dependencies specified in your go.mod file.

### Step 2: Install ginkgo Testing Tool

Next, you need to install the ginkgo testing tool. ginkgo is a popular testing framework for Go. You can install it by running:
```bash
go install github.com/onsi/ginkgo/v2/ginkgo@latest
```
This command will install the latest version of `ginkgo`.

### Step 3: Run All Tests

Once you have ginkgo installed, you can run all the tests in your project using:
```bash
ginkgo ./...
```
This command will execute all the tests in your project directory and its subdirectories.

### Step 4: Build CLI
```bash
go build
```
