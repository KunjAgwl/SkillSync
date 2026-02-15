import { Platform } from 'react-native';
import { storage } from './storage';

// Replace this IP with your computer's LAN IP (run `ipconfig` to find it)
const LAN_IP = '172.20.138.156';

const API_URL = Platform.select({
    android: `http://${LAN_IP}:8080`,
    ios: `http://${LAN_IP}:8080`,
    default: 'http://localhost:8080', // web
});

async function getAuthHeaders(): Promise<Record<string, string>> {
    const token = await storage.getItem('user_token');
    if (!token) return {};
    return {
        'Authorization': `Bearer ${token}`,
    };
}

async function handleResponse(response: Response) {
    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || errorData.error || `Request failed (${response.status})`);
    }
    return response.json();
}

export const api = {
    async register(email: string, password: string) {
        const response = await fetch(`${API_URL}/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });
        return handleResponse(response);
    },

    async login(email: string, password: string) {
        const response = await fetch(`${API_URL}/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });
        return handleResponse(response);
    },

    async uploadResume(fileUri: string, fileName: string, fileType: string) {
        const headers = await getAuthHeaders();
        const formData = new FormData();

        // React Native's FormData handling requires specific structure for files
        formData.append('resume', {
            uri: fileUri,
            name: fileName,
            type: fileType,
        } as any);

        const response = await fetch(`${API_URL}/upload-resume`, {
            method: 'POST',
            headers: {
                ...headers,
                // Do NOT set Content-Type manually for multipart — 
                // fetch will auto-set it with the correct boundary
            },
            body: formData,
        });

        return handleResponse(response);
    },

    async getMatches(userId: number) {
        const headers = await getAuthHeaders();
        const response = await fetch(`${API_URL}/matches?user_id=${userId}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                ...headers,
            },
        });
        return handleResponse(response);
    },
};
