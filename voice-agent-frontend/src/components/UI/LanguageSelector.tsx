import React from 'react';
import { useLanguageStore, Language, translations } from '../../utils/i18n';
import { Globe } from 'lucide-react';

interface LanguageSelectorProps {
    className?: string;
}

const languageNames: Record<Language, string> = {
    en: 'English',
    es: 'Español',
    fr: 'Français',
    de: 'Deutsch',
    hi: 'हिन्दी',
    ja: '日本語',
    zh: '中文',
};

export const LanguageSelector: React.FC<LanguageSelectorProps> = ({ className = '' }) => {
    const { language, setLanguage } = useLanguageStore();

    return (
        <div className={`flex items-center gap-2 ${className}`}>
            <Globe className="w-4 h-4 text-gray-500" />
            <select
                value={language}
                onChange={(e) => setLanguage(e.target.value as Language)}
                className="text-sm border border-gray-300 rounded-lg px-3 py-2 bg-white hover:bg-gray-50 transition-colors focus:outline-none focus:ring-2 focus:ring-agent-primary"
            >
                {Object.entries(languageNames).map(([code, name]) => (
                    <option key={code} value={code}>
                        {name}
                    </option>
                ))}
            </select>
        </div>
    );
};

export default LanguageSelector;
