go tool pprof cpu.prof
go tool pprof -dot cpu.prof > cpu.dot
go tool pprof -dot mem.prof > mem.dot

