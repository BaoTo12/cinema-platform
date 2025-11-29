'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { TransportProvider } from '@connectrpc/connect-query';
import { createConnectTransport } from '@connectrpc/connect-web';
import { useState } from 'react';

export function Providers({ children }: { children: React.ReactNode }) {
    const [queryClient] = useState(
        () =>
            new QueryClient({
                defaultOptions: {
                    queries: {
                        staleTime: 60 * 1000, // 1 minute
                    },
                },
            })
    );

    const [transport] = useState(() =>
        createConnectTransport({
            baseUrl: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:5000',
            // Add auth interceptor
            interceptors: [
                (next) => async (req) => {
                    const token = localStorage.getItem('accessToken');
                    if (token) {
                        req.header.set('Authorization', `Bearer ${token}`);
                    }
                    return await next(req);
                },
            ],
        })
    );

    return (
        <TransportProvider transport={transport}>
            <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
        </TransportProvider>
    );
}
