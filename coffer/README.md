# coffer [![Build status](https://ci.appveyor.com/api/projects/status/gpka5vixsysfbcjs?svg=true)](https://ci.appveyor.com/project/dpull/coffer)

File encryption utility. It can encrypt files in a directory, and is easily extensible.

[Download Windows x64 ZIP](https://ci.appveyor.com/project/dpull/coffer/build/artifacts)

## Usage
```
config.json
{
    "http_addr": "localhost:8964",      // http service port
    "folder": "./temp",                 // local directory to be mapped
    "fs_type": "xor",                   // filesystem type
    "fs_param": {                       // filesystem parameters, i.e. security key
        "key": "qwertyuiop[]" 
    }
}
```

`./coffer [-c config] [-l log]`
* `-c config`, configuration file to be used, default is `config.json`
* `-l log` log file, default is `coffer.log`

After the utility is launched, use `explorer` under Windows, `Finder` under MacOS, to map the URL to local directory. For example, map `http://localhost:8964` as configured like above. Then the filesystem can be used.

## performance
```
/Volumes % sync; dd if=/dev/zero of=/Volumes/localhost/tempfile bs=1M count=1024; sync
1024+0 records in
1024+0 records out
1073741824 bytes transferred in 1.640524 secs (654511500 bytes/sec)
/Volumes % sync; dd if=/Volumes/localhost/tempfile of=/dev/null bs=1M count=1024; sync
1024+0 records in
1024+0 records out
1073741824 bytes transferred in 0.105351 secs (10192042069 bytes/sec)
```

```
/Volumes % sync; dd if=/dev/zero of=/Users/x/Documents/tempfile bs=1M count=1024; sync
1024+0 records in
1024+0 records out
1073741824 bytes transferred in 0.157545 secs (6815461132 bytes/sec)
/Volumes % sync; dd if=/Users/x/Documents/tempfile of=/dev/null bs=1M count=1024; sync
1024+0 records in
1024+0 records out
1073741824 bytes transferred in 0.075648 secs (14193922166 bytes/sec)
```