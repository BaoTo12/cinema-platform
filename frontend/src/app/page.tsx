import Link from 'next/link';
import { Film, Ticket } from 'lucide-react';

export default function HomePage() {
    return (
        <div className="min-h-screen">
            {/* Navigation */}
            <nav className="border-b border-dark-lighter">
                <div className="container mx-auto px-4 py-4 flex items-center justify-between">
                    <Link href="/" className="flex items-center gap-2">
                        <Film className="w-8 h-8 text-primary-500" />
                        <span className="text-2xl font-display font-bold">CinemaOS</span>
                    </Link>
                    <div className="flex items-center gap-4">
                        <Link href="/movies" className="hover:text-primary-500 transition-colors">
                            Movies
                        </Link>
                        <Link href="/login" className="btn-secondary">
                            Login
                        </Link>
                    </div>
                </div>
            </nav>

            {/* Hero Section */}
            <section className="relative py-20 overflow-hidden">
                <div className="absolute inset-0 bg-gradient-radial from-primary-900/20 to-transparent"></div>
                <div className="container mx-auto px-4 relative z-10">
                    <div className="max-w-4xl mx-auto text-center">
                        <h1 className="text-6xl font-display font-bold mb-6 bg-gradient-to-r from-primary-500 to-accent-gold bg-clip-text text-transparent">
                            Experience Cinema Like Never Before
                        </h1>
                        <p className="text-xl text-gray-400 mb-8">
                            Book your tickets instantly with real-time seat selection and dynamic pricing
                        </p>
                        <div className="flex gap-4 justify-center">
                            <Link href="/movies" className="btn-primary text-lg px-8 py-3">
                                <Ticket className="inline-block w-5 h-5 mr-2" />
                                Book Now
                            </Link>
                            <Link href="/about" className="btn-secondary text-lg px-8 py-3">
                                Learn More
                            </Link>
                        </div>
                    </div>
                </div>
            </section>

            {/* Features */}
            <section className="py-16">
                <div className="container mx-auto px-4">
                    <h2 className="text-3xl font-display font-bold text-center mb-12">
                        Why Choose CinemaOS?
                    </h2>
                    <div className="grid md:grid-cols-3 gap-8">
                        <div className="card text-center">
                            <div className="w-16 h-16 bg-primary-500/10 rounded-full flex items-center justify-center mx-auto mb-4">
                                <Ticket className="w-8 h-8 text-primary-500" />
                            </div>
                            <h3 className="text-xl font-semibold mb-2">Real-time Booking</h3>
                            <p className="text-gray-400">
                                See seat availability in real-time and book instantly
                            </p>
                        </div>
                        <div className="card text-center">
                            <div className="w-16 h-16 bg-accent-gold/10 rounded-full flex items-center justify-center mx-auto mb-4">
                                <Film className="w-8 h-8 text-accent-gold" />
                            </div>
                            <h3 className="text-xl font-semibold mb-2">Latest Movies</h3>
                            <p className="text-gray-400">
                                Watch the newest blockbusters in premium quality
                            </p>
                        </div>
                        <div className="card text-center">
                            <div className="w-16 h-16 bg-accent-blue/10 rounded-full flex items-center justify-center mx-auto mb-4">
                                <span className="text-2xl">💰</span>
                            </div>
                            <h3 className="text-xl font-semibold mb-2">Dynamic Pricing</h3>
                            <p className="text-gray-400">
                                Get the best deals with our smart pricing system
                            </p>
                        </div>
                    </div>
                </div>
            </section>

            {/* CTA Section */}
            <section className="py-16 bg-gradient-to-r from-primary-900/20 to-accent-gold/20">
                <div className="container mx-auto px-4 text-center">
                    <h2 className="text-4xl font-display font-bold mb-4">
                        Ready to Watch?
                    </h2>
                    <p className="text-gray-300 mb-8">
                        Browse our collection of movies and book your perfect seat
                    </p>
                    <Link href="/movies" className="btn-primary text-lg px-8 py-3">
                        Explore Movies
                    </Link>
                </div>
            </section>

            {/* Footer */}
            <footer className="border-t border-dark-lighter py-8">
                <div className="container mx-auto px-4 text-center text-gray-400">
                    <p>© 2024 CinemaOS. Built with Golang + Next.js + Connect RPC</p>
                </div>
            </footer>
        </div>
    );
}
