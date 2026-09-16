// fetchAuth.js
export const fetchWithAuth = async (url, options = {}) => {
  const isFormData = options.body instanceof FormData;

  const headers = {
    ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
    ...options.headers
  };

  const response = await fetch(url, {
    ...options,
    headers,
    credentials: 'include' // Sends the HttpOnly cookie
  });

  return response;
};