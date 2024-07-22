import { createApiClient } from "./apiclient"

const apiclient = createApiClient("http://localhost:8080")

apiclient.beforeRequest((config: RequestInit) => {
  if (config?.headers) {
    config.headers["Authorization"] = "lol"
  }
})

const response = await apiclient.main.HelloWorld({ name: "asdfa" })

console.log(response.data, response.error)
