# Nginless

## Config file

```
{
  "log": {
    "path": "/tmp/nginless.log",
    "maxSize": 1000,
    "maxBackups": 20,
    "maxAge": 14
  }
}
```

## Run

```
nginless -p ${port} -r ${router_yaml_file} -a ${action_script_folder}
```
