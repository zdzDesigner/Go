import { spawn } from "child_process"

const sleep = (delay) => new Promise((resolve, _) => setTimeout(resolve, delay))

const sp = spawn("./stdio", [], {
  stdio: ["pipe", "pipe", "pipe"], // 启用 stdin/stdout/stderr
})

// 监听 C 程序的 stdout 输出
sp.stdout.on("data", (data) => {
  console.log("sub receive:", data.toString())
})
sp.on("spawn", async () => {
  console.log("-------- ok, sub send ... xxxx")
  const data = { name: "probe", data: "1111" }
  const res = sp.stdin.write(`${JSON.stringify(data)}\n`, (...args) => {
    console.log({ args })
  })
  console.log({ res })
})

setTimeout(() => {
  const data = { name: "peq", data: "99999999999999999999999999" }
  sp.stdin.write(`${JSON.stringify(data)}\n`)
}, 8000)

sp.stderr.on("data", (data) => {
  console.log("--:", data.toString())
})
sp.on("close", (code) => {
  console.log("--------- close")
})
sp.on("error", (err) => {
  console.log("--------- error")
})
