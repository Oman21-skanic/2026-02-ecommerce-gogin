export const API_BASE_URL =
  import.meta.env.API_BASE_URL?.replace(/\/$/, "") || "http://localhost:8080";

export const PUBLIC_SITE_URL =
  import.meta.env.PUBLIC_SITE_URL?.replace(/\/$/, "") || "http://localhost:4321";
