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
Sync add new value to object and keep old value. Usually use when updating fields of record.
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
