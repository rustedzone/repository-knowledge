import type {NextApiRequest, NextApiResponse} from "next";

export default async function handler(request: NextApiRequest, response: NextApiResponse) {
  const path = Array.isArray(request.query.paths) ? request.query.paths.join("/") : "";
  const upstream = await fetch(`${process.env.API_BASE_URL}/${path}`, {
    method: request.method,
    headers: {
      "content-type": "application/json",
      cookie: request.headers.cookie ?? "",
    },
    body: request.method === "GET" ? undefined : JSON.stringify(request.body),
  });

  const payload = await upstream.json();
  response.status(upstream.status).json(payload);
}
