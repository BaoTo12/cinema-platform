'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Film, X } from 'lucide-react';

export default function BookingPage({ params }: { params: { showtimeId: string } }) {
    const router = useRouter();
    const [selectedSeats, setSelectedSeats] = useState<string[]>([]);

    // Mock seat map - would come from Connect RPC
    const seatRows = [
        {
            label: 'A', seats: Array.from({ length: 10 }, (_, i) => ({
                id: `A${i + 1}`,
                number: i + 1,
                type: 'STANDARD' as const,
                status: i < 2 ? 'BOOKED' as const : 'AVAILABLE' as const,
                price: 10
            }))
        },
        {
            label: 'B', seats: Array.from({ length: 10 }, (_, i) => ({
                id: `B${i + 1}`,
                number: i + 1,
                type: 'PREMIUM' as const,
                status: 'AVAILABLE' as const,
                price: 13
            }))
        },
    ];

    const toggleSeat = (seatId: string) => {
        setSelectedSeats(prev =>
            prev.includes(seatId)
                ? prev.filter(id => id !== seatId)
                : [...prev, seatId]
        );
    };

    const getSeatColor = (status: string, isSelected: boolean) => {
        if (isSelected) return 'bg-primary-500';
        if (status === 'BOOKED') return 'bg-gray-600 cursor-not-allowed';
        if (status === 'LOCKED') return 'bg-yellow-600';
        return 'bg-gray-400 hover:bg-primary-400';
    };

    const totalPrice = selectedSeats.length * 10; // Simplified calculation

    return (
        <div className="min-h-screen">
            {/* Nav */}
            <nav className="border-b border-dark-lighter">
                <div className="container mx-auto px-4 py-4 flex items-center justify-between">
                    <Link href="/" className="flex items-center gap-2">
                        <Film className="w-8 h-8 text-primary-500" />
                        <span className="text-2xl font-display font-bold">CinemaOS</span>
                    </Link>
                </div>
            </nav>

            <div className="container mx-auto px-4 py-8">
                <div className="max-w-5xl mx-auto">
                    <h1 className="text-3xl font-display font-bold mb-2">Select Your Seats</h1>
                    <p className="text-gray-400 mb-8">The Dark Knight • 20:00 • Screen 1</p>

                    {/* Screen */}
                    <div className="mb-12">
                        <div className="bg-gradient-to-b from-gray-700 to-dark h-2 rounded-t-full mb-2"></div>
                        <p className="text-center text-gray-400 text-sm">SCREEN</p>
                    </div>

                    {/* Seat Map */}
                    <div className="space-y-4 mb-8">
                        {seatRows.map((row) => (
                            <div key={row.label} className="flex items-center gap-4">
                                <span className="w-8 text-center font-bold text-gray-400">{row.label}</span>
                                <div className="flex-1 flex justify-center gap-2">
                                    {row.seats.map((seat) => {
                                        const isSelected = selectedSeats.includes(seat.id);
                                        const isAvailable = seat.status === 'AVAILABLE';

                                        return (
                                            <button
                                                key={seat.id}
                                                onClick={() => isAvailable && toggleSeat(seat.id)}
                                                disabled={!isAvailable}
                                                className={`w-8 h-8 rounded-t-lg transition-colors ${getSeatColor(seat.status, isSelected)}`}
                                                title={`${row.label}${seat.number} - ${seat.type} - $${seat.price}`}
                                            >
                                                {isSelected && <span className="text-xs">✓</span>}
                                            </button>
                                        );
                                    })}
                                </div>
                            </div>
                        ))}
                    </div>

                    {/* Legend */}
                    <div className="flex justify-center gap-6 mb-8 text-sm">
                        <div className="flex items-center gap-2">
                            <div className="w-6 h-6 bg-gray-400 rounded-t-lg"></div>
                            <span>Available</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <div className="w-6 h-6 bg-primary-500 rounded-t-lg"></div>
                            <span>Selected</span>
                        </div>
                        <div className="flex items-center gap-2">
                            <div className="w-6 h-6 bg-gray-600 rounded-t-lg"></div>
                            <span>Booked</span>
                        </div>
                    </div>

                    {/* Checkout Bar */}
                    {selectedSeats.length > 0 && (
                        <div className="fixed bottom-0 left-0 right-0 bg-dark-light border-t border-dark-lighter p-4">
                            <div className="container mx-auto flex items-center justify-between">
                                <div>
                                    <p className="text-sm text-gray-400">Selected Seats</p>
                                    <p className="font-bold">{selectedSeats.join(', ')}</p>
                                </div>
                                <div className="flex items-center gap-4">
                                    <div className="text-right">
                                        <p className="text-sm text-gray-400">Total</p>
                                        <p className="text-2xl font-bold text-primary-500">${totalPrice.toFixed(2)}</p>
                                    </div>
                                    <button className="btn-primary">
                                        Continue to Payment
                                    </button>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
