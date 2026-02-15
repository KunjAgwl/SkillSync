import React, { createContext, useContext, useEffect, useState } from 'react';
import { storage } from '@/services/storage';
import { api } from '@/services/api';

type AuthContextType = {
    user: any | null;
    isLoading: boolean;
    signIn: (email: string, password: string) => Promise<void>;
    signUp: (email: string, password: string) => Promise<void>;
    signOut: () => Promise<void>;
    session?: string | null;
};

const AuthContext = createContext<AuthContextType>({
    user: null,
    isLoading: true,
    signIn: async () => { },
    signUp: async () => { },
    signOut: async () => { },
    session: null,
});

export function useAuth() {
    return useContext(AuthContext);
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [user, setUser] = useState<any | null>(null);
    const [session, setSession] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        // Check for stored session on mount
        const loadSession = async () => {
            try {
                const token = await storage.getItem('user_token');
                const userData = await storage.getItem('user_data');

                if (token && userData) {
                    setSession(token);
                    setUser(JSON.parse(userData));
                }
            } catch (e) {
                console.error('Failed to load session', e);
            } finally {
                setIsLoading(false);
            }
        };
        loadSession();
    }, []);

    const signIn = async (email: string, password: string) => {
        setIsLoading(true);
        try {
            const response = await api.login(email, password);
            // Assuming response has structure: { token: string, user: object }
            // Adjust based on your actual API response in services/api.ts

            setSession(response.token);
            setUser(response.user);

            await storage.setItem('user_token', response.token);
            await storage.setItem('user_data', JSON.stringify(response.user));
        } catch (e) {
            throw e;
        } finally {
            setIsLoading(false);
        }
    };

    const signUp = async (email: string, password: string) => {
        setIsLoading(true);
        try {
            // Register usually returns the new user, sometimes a token too.
            // If it doesn't return a token, we might need to login immediately after.
            // For now, let's assume it might just create the user.
            const response = await api.register(email, password);

            // If your API automatically logs in after register, handle it here.
            // If not, the user will be redirected to login (or we call signIn here).
            // Let's assume we need to sign in after register for now, or just let the UI handle it.
            // For this implementation, let's just return and let the UI redirect to Login.
            if (response.token) {
                setSession(response.token);
                setUser(response.user);
                await storage.setItem('user_token', response.token);
                await storage.setItem('user_data', JSON.stringify(response.user));
            }
        } catch (e) {
            throw e;
        } finally {
            setIsLoading(false);
        }
    }

    const signOut = async () => {
        setIsLoading(true);
        try {
            await storage.deleteItem('user_token');
            await storage.deleteItem('user_data');
            setSession(null);
            setUser(null);
        } catch (e) {
            console.error(e);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <AuthContext.Provider
            value={{
                signIn,
                signUp,
                signOut,
                user,
                session,
                isLoading,
            }}>
            {children}
        </AuthContext.Provider>
    );
}
