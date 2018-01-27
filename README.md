# IoC(DJ) framework for Go
This is a lightweight Inversion of Control (Dependency Injection) framework for Go. It's very convenient to register concrete value, singleton or transient prototype into the container, and resolve the value somewhere else with all dependencies automatically injected. 
### Name and Alias  
Each registry will return its name as identifier, which is composed of type's package path and name.
1. How to get the name  
`ioc.NameOf(v)` will returns v's name
1. Concrete value's name  
`ioc.NameOf(&Rectangle{})` is `github.com/gopub/ioc_test/*Rectangle`  
`ioc.NameOf(Rectangle{})` is `github.com/gopub/ioc_test/Rectangle`  
2. Interface's name  
Must pass pointer to interface in order to get its name  
Error:  
`ioc.NameOf(Shape(nil))` is `nil`   
Correct:  
`ioc.NameOf((*Shape)(nil))` is `github.com/gopub/ioc_test/Shape`
3. Alias for name  

    ```
    name := ioc.RegisterSingleton(&Rectangle{})
    
    //Add "MyRectangle" as alias for &Rectangle{}
    ioc.RegisterAliases(name, "MyRectangle")
    
    //r1 and r2 is equal
    r1 := ioc.Resolve(name)
    r2 := ioc.Resolve("MyRectangle")
    ```
### Register and Resolve
1. Concrete value  

    ```
    //register
    ioc.RegisterValue("key", "123456")
    ioc.RegisterValue("db", db)
    
    //fetch value in somewhere else
    key := ioc.Resolve("key").(string)
    db := ioc.Resolve("db").(*sql.DB)
    ```
2. Singleton prototype

    RegisterSingleton ensures only one value will be created no matter how many times it is resolved.
    ```
    //register
    ioc.RegisterSingleton(&LoginService{})
    
    //fetch value in somewhere else
    //s1 is the same with s2
    s1 := ioc.Resolve(ioc.NameOf(&LoginService)).(*LoginService)
    s2 := ioc.Resolve(ioc.NameOf(&LoginService)).(*LoginService)
    ```
3. Transient prototype  
    Use RegisterTransient to register a type of which many values will be created. 
    ``` 
    ioc.RegisterTransient(&Rectangle{})
    
    //r1 and r2 are different values.
    r1 := ioc.Resolve(ioc.NameOf(&Rectangle)).(*Rectangle)
    r2 := ioc.Resolve(ioc.NameOf(&Rectangle)).(*Rectangle)
    ```
4. Prototype for interface
    It's very common to bind a concrete type to an interface type. To support this scenario, register interface's name as an alias of concrete type's name. 
    ```
    //Rectange implements Shape interface
    rectName := ioc.RegisterSingleton(&Rectangle{})
    shapeName := ioc.NameOf((*Shape)(nil))
    ioc.RegisterAliases(rectName, shapeName)
    
    //Create &Rectange{} as Shape
    s := ioc.Resolve(shapeName).(Shape)
    fmt.Print(s.Area())                    
    ```
5. Timing
    The type and its dependencies must be registered before resolve. It's good time to do registry operations before main function executes.
    1. In package's init() function
    ``` 
        func init() {
            ioc.Resolve(ioc.NameOf(&LoginService))
        }
    ```
    2. Declare global variable to trigger registry operation
    ``` 
        //Declare just ahead of type definition
        var _ = ioc.Resolve(ioc.NameOf(&LoginService))
        type LoginService struct {
        
        }
    ```
### Declare dependencies
Use tag `inject` to declare dependencies. Field name must start with an uppercase character in case it will cause unexported panic error.
``` 
var _ = ioc.RegisterSingleton(&LoginController{})

type LoginController struct {
    //depend value with name: page_title
    PageTitle    string         `inject:"page_title"`
    
    //depend value with name: ioc.NameOf((*LoginService)(nil))
    LoginService *LoginService  `inject:""`
}

//somewhere else
controller := ioc.Resolve(&LoginController{}).(*LoginController)
//...
```