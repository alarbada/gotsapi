import { createApiClient } from "./apiclient"

const apiclient = createApiClient("http://localhost:8080")

apiclient.beforeRequest((config: RequestInit) => {
  if (config?.headers) {
    config.headers["Authorization"] = "lol"
  }
})

const response = await apiclient.main.ExampleHandler1({
  name: "username",
  users: [
    {
      name: "username",
      age: 1000,
    },
  ],
})

console.log(response.error)
