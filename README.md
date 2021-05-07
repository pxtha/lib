:collision: Falling from starts
# Common functions for golang coder
- Leave a Star if you were here!

# Table of Content
 - [Usage](#usage)
 - [Sync](#sync)
 - [Validation](#Validation)
 - [Send Rest API to other serives](#send-rest-api)
 - [Map to struct](#map-to-struct)

# Usage

```bash
go get github.com/praslar/lib@latest
```

# Sync 
Sync add only changed values to object. Usually use when updating fields of record.
```golang

type UpdateUserRequest struct {
    Name *string
}

type User struct {
    ID   string
    Name string
    Age  int
    School string
    ...
}

func UpdateUser(req UpdateUserRequest, id string) (res User) {
    user := model.GetUser(iD)
    
    common.Sync(req,&user)
    // user will have new value of req.Name and keep others values unchange
    
    _ = model.Update(user)
    ...
}

```
# Validation
Valid struct and return error if there is any.
```golang
req := struct {
    Name string `valid:"Required"`
    Age  int     `valid:"Required"`
}{"Praslar"}

if err := common.CheckRequireValid(req); err != nil {
		return nil, err
}

// this will return error that required Age value in req object
```
