'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Film, Calendar, Clock, MapPin, ChevronLeft } from 'lucide-react';

export default function MovieDetailPage({ params }: { params: { id: string } }) {
    const [selectedDate, setSelectedDate] = useState('2024-01-15');

    // Mock data - would come from Connect RPC
    const movie = {
        id: params.id,
        title: 'The Dark Knight',
        description: 'When the menace known as the Joker wreaks havoc on Gotham...',
        posterUrl: 'https://via.placeholder.com/300x450',
        backdropUrl: 'https://via.placeholder.com/1200x400',
        rating: 'PG-13',
        duration: 152,
        genres: ['Action', 'Crime', 'Drama'],
    };

    const showtimes = [
        { id: '1', time: '14:00', screen: 'Screen 1', availableSeats: 45 },
        { id: '2', time: '17:30', screen: 'Screen 2', availableSeats: 12 },
        { id: '3', time: '20:00', screen: 'Screen 1', availableSeats: 78 },
    ];

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

            {/* Hero */}
            <div className="relative h-96 overflow-hidden">
                <div
                    className="absolute inset-0 bg-cover bg-center"
                    style={{ backgroundImage: `url(${movie.backdropUrl})` }}
                >
                    <div className="absolute inset-0 bg-gradient-to-t from-dark via-dark/80 to-dark/40"></div>
                </div>
                <div className="container mx-auto px-4 h-full flex items-end pb-8 relative z-10">
                    <Link href="/movies" className="absolute top-4 left-4 flex items-center gap-2 text-gray-300 hover:text-white">
                        <ChevronLeft className="w-5 h-5" />
                        Back to Movies
                    </Link>
                    <div className="flex gap-6">
                        <img src={movie.posterUrl} alt={movie.title} className="w-48 rounded-lg shadow-2xl" />
                        <div>
                            <h1 className="text-5xl font-display font-bold mb-2">{movie.title}</h1>
                            <div className="flex items-center gap-4 text-gray-300 mb-4">
                                <span className="px-3 py-1 bg-dark-lighter rounded">{movie.rating}</span>
                                <span className="flex items-center gap-1">
                                    <Clock className="w-4 h-4" />
                                    {movie.duration} min
                                </span>
                                <span>{movie.genres.join(', ')}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Content */}
            <div className="container mx-auto px-4 py-8">
                <div className="grid lg:grid-cols-3 gap-8">
                    {/* Description */}
                    <div className="lg:col-span-2">
                        <h2 className="text-2xl font-bold mb-4">Synopsis</h2>
                        <p className="text-gray-300 leading-relaxed mb-8">{movie.description}</p>

                        {/* Showtimes */}
                        <h2 className="text-2xl font-bold mb-4">Select Showtime</h2>
                        <div className="flex gap-2 mb-6">
                            {['Jan 15', 'Jan 16', 'Jan 17'].map((date) => (
                                <button
                                    key={date}
                                    className="px-4 py-2 rounded-lg bg-dark-lighter hover:bg-dark text-sm font-medium"
                                >
                                    {date}
                                </button>
                            ))}
                        </div>

                        <div className="space-y-3">
                            {showtimes.map((showtime) => (
                                <Link
                                    key={showtime.id}
                                    href={`/booking/${showtime.id}`}
                                    className="card flex items-center justify-between hover:border-primary-500 transition-colors"
                                >
                                    <div className="flex items-center gap-4">
                                        <div className="text-2xl font-bold text-primary-500">{showtime.time}</div>
                                        <div>
                                            <div className="font-medium">{showtime.screen}</div>
                                            <div className="text-sm text-gray-400">{showtime.availableSeats} seats available</div>
                                        </div>
                                    </div>
                                    <div className="btn-primary">
                                        Select
                                    </div>
                                </Link>
                            ))}
                        </div>
                    </div>

                    {/* Sidebar */}
                    <div>
                        <div className="card sticky top-4">
                            <h3 className="font-bold mb-4">Cinema Location</h3>
                            <div className="flex items-start gap-2 text-gray-300">
                                <MapPin className="w-5 h-5 mt-1 flex-shrink-0" />
                                <div>
                                    <p>Downtown Cinema</p>
                                    <p className="text-sm text-gray-400">123 Main St, City</p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
