# Luajit

Go 库：离线解析 LuaJIT 2.1 dump（标准 `1B 4C 4A 02`、迷你世界 `1B 4C 4A 90`），并反编译成 Lua。

```text
github.com/PaperCodeDevs/Luajit
```

```text
.
luajit.go          对外 API
parse/             dump 头、proto、kgc/knum、扫描
op/                标准 opcode 名 + 迷你世界编号对照
lua/               反汇编 / 反编译 / 批量
cmd/ljdump         进程入口
```

```text
ljdump <file|dir> [out-dir]
```
