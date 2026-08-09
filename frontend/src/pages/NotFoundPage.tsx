import { ArrowLeft, MapPinOff } from 'lucide-react'
import { Link } from 'react-router'
export function NotFoundPage() { return <div className="page-container not-found"><MapPinOff aria-hidden="true" /><p className="eyebrow">404 · Route not found</p><h1>This minute is outside the horizon.</h1><p>The page you requested is not part of the current TimeTrap scenario.</p><Link className="button primary" to="/"><ArrowLeft aria-hidden="true" /> Return home</Link></div> }
