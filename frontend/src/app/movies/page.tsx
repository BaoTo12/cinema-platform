'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Film, Search } from 'lucide-react';

export default function MoviesPage() {
    const [search, setSearch] = useState('');

    // In production, this would use Connect RPC to fetch movies
    const movies = [
        {
            id: '1',
            title: 'The Dark Knight',
            posterUrl: 'https://via.placeholder.com/300x450',
            rating: 'PG-13',
            genres: ['Action', 'Drama'],
            duration: 152,
        },
        // More movies...
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
                    <div className="flex items-center gap-4">
                        <Link href="/movies" className="text-primary-500">Movies</Link>
                        <Link href="/my-bookings">My Bookings</Link>
                        <Link href="/login" className="btn-secondary">Login</Link>
                    </div>
                </div>
            </nav>

            <div className="container mx-auto px-4 py-8">
                {/* Header */}
                <div className="mb-8">
                    <h1 className="text-4xl font-display font-bold mb-4">Now Showing</h1>
                    <div className="relative max-w-md">
                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
                        <input
                            type="text"
                            placeholder="Search movies..."
                            className="input pl-10"
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                        />
                    </div>
                </div>

                {/* Movies Grid */}
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                    {movies.map((movie) => (
                        <Link key={movie.id} href={`/movies/${movie.id}`} className="group">
                            <div className="card p-0 overflow-hidden hover:border-primary-500 transition-all duration-300">
                                <div className="aspect-[2/3] bg-dark-lighter relative overflow-hidden">
                                    <img
                                        src={movie.posterUrl}
                                        alt={movie.title}
                                        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                                    />
                                </div>
                                <div className="p-4">
                                    <h3 className="font-semibold text-lg mb-1 group-hover:text-primary-500 transition-colors">
                                        {movie.title}
                                    </h3>
                                    <div className="flex items-center gap-2 text-sm text-gray-400">
                                        <span className="px-2 py-0.5 bg-dark-lighter rounded text-xs">
                                            {movie.rating}
                                        </span>
                                        <span>{movie.duration} min</span>
                                    </div>
                                    <div className="mt-2 flex flex-wrap gap-1">
                                        {movie.genres.slice(0, 2).map((genre) => (
                                            <span key={genre} className="text-xs text-primary-400">
                                                {genre}
                                            </span>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        </Link>
                    ))}
                </div>
            </div>
        </div>
    );
}
