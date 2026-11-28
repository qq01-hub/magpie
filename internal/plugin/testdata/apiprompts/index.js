// A plugin whose "api" ways ask questions, as inkstone's and Devin's
// session-token way do: one whose authorize turns the answers into the key,
// and one with no authorize, whose key magpie asks after them.
export const ApiPromptsPlugin = async () => ({
  auth: {
    provider: "apiprompts",
    methods: [
      // 0: the key is a question's answer; authorize keeps it as the key
      {
        type: "api",
        label: "Session token",
        prompts: [
          { type: "select", key: "site", message: "Site", options: [{ label: "Global", value: "global" }, { label: "EU", value: "eu" }] },
          { type: "password", key: "token", message: "Session token" },
          { type: "text", key: "note", message: "Console cookie" },
        ],
        authorize: async (inputs) => ({ type: "success", key: inputs.token, metadata: { email: "who@apiprompts.invalid" } }),
      },
      // 1: no authorize: the answers are the account's metadata, the key
      // asked beside them (OpenCode's dialog does the same)
      {
        type: "api",
        label: "Site key",
        prompts: [{ type: "text", key: "tenant", message: "Tenant" }],
      },
      // 2: an authorize and no questions: the key is magpie's to ask
      { type: "api", label: "Checked key", authorize: async () => ({ type: "success", metadata: { checked: "yes" } }) },
      // #6064: PackyAPI 0.2.1 asks only nonsecret account and endpoint choices
      {
        type: "api",
        label: "System token",
        prompts: [
          { type: "text", key: "userID", message: "User ID" },
          { type: "select", key: "endpoint", message: "Endpoint", options: [{ label: "SLB", value: "slb" }] },
        ],
        authorize: async (inputs) => ({ type: "success", metadata: { mode: "token", userID: inputs.userID, endpoint: inputs.endpoint || "slb", email: `packyapi-${inputs.userID}` } }),
      },
      {
        type: "api",
        label: "API keys",
        prompts: [{ type: "select", key: "endpoint", message: "Endpoint", options: [{ label: "SLB", value: "slb" }] }],
        authorize: async (inputs) => ({ type: "success", metadata: { mode: "keys", endpoint: inputs.endpoint || "slb" } }),
      },
    ],
  },
})
