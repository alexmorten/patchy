const API_BASE_URL = 'http://localhost:7788/api';

interface ApiError extends Error {
  status?: number;
}

export const search = async (query: string) => {
  try {
    const response = await fetch(`${API_BASE_URL}/search?q=${encodeURIComponent(query)}`, {
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      const error = new Error('Search request failed') as ApiError;
      error.status = response.status;
      throw error;
    }

    return await response.json();
  } catch (error) {
    console.error('Search error:', error);
    throw error;
  }
};

export default API_BASE_URL; 