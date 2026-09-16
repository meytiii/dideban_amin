'use client';

import React, { useState } from 'react';
import { ChevronDown, HelpCircle, Shield, FileText, Search, Paperclip } from 'lucide-react';

interface AccordionItemProps {
  value: string;
  title: string;
  content: string;
  icon?: React.ReactNode;
  isOpen: boolean;
  onToggle: (value: string) => void;
}

const AccordionItem = ({ value, title, content, icon, isOpen, onToggle }: AccordionItemProps) => {
  return (
    <div 
      className={`border rounded-2xl px-5 transition-all duration-300 ${
        isOpen
          ? 'bg-slate-900/80 border-blue-500/40 shadow-lg shadow-blue-500/5'
          : 'bg-slate-900/40 border-slate-800/80 hover:border-slate-700/80'
      } backdrop-blur-md mb-3`}
    >
      <button
        type="button"
        onClick={() => onToggle(value)}
        className="w-full text-right flex justify-between items-center py-4.5 transition-colors cursor-pointer group"
      >
        <div className="flex items-center gap-3 pr-1 min-w-0">
          {icon && (
            <div className={`p-2 rounded-xl border transition-colors ${
              isOpen 
                ? 'bg-blue-500/20 border-blue-500/40 text-blue-400' 
                : 'bg-slate-800/60 border-slate-700/50 text-slate-400 group-hover:text-slate-300'
            }`}>
              {icon}
            </div>
          )}
          <span className={`text-[14.5px] font-semibold transition-colors ${
            isOpen ? 'text-white' : 'text-slate-200 group-hover:text-white'
          }`}>
            {title}
          </span>
        </div>

        <ChevronDown 
          className={`w-4.5 h-4.5 flex-shrink-0 text-slate-400 transition-transform duration-300 ${
            isOpen ? 'transform rotate-180 text-blue-400' : 'group-hover:text-slate-200'
          }`}
        />
      </button>

      <div 
        className={`overflow-hidden transition-all duration-300 ${
          isOpen ? 'max-h-96 opacity-100' : 'max-h-0 opacity-0'
        }`}
      >
        <div className="pb-4.5 pt-1 text-[13px] leading-relaxed text-slate-300 border-t border-slate-800/60 mt-1">
          {content}
        </div>
      </div>
    </div>
  );
};

export function FAQAccordion() {
  const [openItem, setOpenItem] = useState<string | null>(null);

  const handleToggle = (value: string) => {
    setOpenItem(openItem === value ? null : value);
  };

  const faqData = [
    {
      value: 'item-1',
      title: 'امنیت و محرمانگی گزارش‌ها چگونه تضمین می‌شود؟',
      icon: <Shield className="w-4 h-4" />,
      content: 'تمامی داده‌ها با پیشرفته‌ترین استانداردهای رمزنگاری سازمانی (AES-256) دریافت شده و تنها کارشناسان مجاز حراست امکان دسترسی به گزارش‌ها را دارند. هیچ‌گونه ردپای هویتی در گزارش‌های ناشناس ذخیره نمی‌شود.'
    },
    {
      value: 'item-2',
      title: 'چه اطلاعاتی به تسریع روند رسیدگی کمک می‌کند؟',
      icon: <FileText className="w-4 h-4" />,
      content: 'ذکر دقیق زمان و مکان وقوع رویداد، نام اشخاص یا واحدهای مرتبط و در صورت امکان پیوست اسناد، تصاویر یا فایل‌های مستند به صورت فشرده، کارشناسان حراست را در ارزیابی سریع و دقیق پرونده یاری می‌کند.'
    },
    {
      value: 'item-3',
      title: 'چگونه می‌توانم نتیجه گزارش خود را پیگیری کنم؟',
      icon: <Search className="w-4 h-4" />,
      content: 'پس از ثبت موفقیت‌آمیز، یک کد پیگیری ۱۶ رقمی اختصاصی به شما ارائه می‌شود. با نگهداری این کد و مراجعه به تب «پیگیری گزارش»، می‌توانید پاسخ رسمی اداره حراست را مشاهده کرده و در صورت لزوم پیام تکمیلی ارسال نمایید.'
    },
    {
      value: 'item-4',
      title: 'فایل‌ها و مستندات مجاز چه مواردی هستند؟',
      icon: <Paperclip className="w-4 h-4" />,
      content: 'ارسال اسناد متنی (PDF)، تصاویر (JPG, PNG) و فایل‌های فشرده (ZIP, RAR) تا سقف ۲۰ مگابایت مجاز است. برای ارسال چند مدرک به صورت همزمان، تجمیع آنها در قالب یک فایل فشرده توصیه می‌شود.'
    }
  ];

  return (
    <div className="w-full relative rounded-3xl overflow-hidden bg-slate-950 p-6 border border-slate-800/80">
      {/* Orthogonal Grid Background */}
      <div 
        className="absolute inset-0 pointer-events-none opacity-20"
        style={{
          backgroundImage: 'linear-gradient(rgba(255,255,255,0.06) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.06) 1px, transparent 1px)',
          backgroundSize: '36px 36px'
        }}
      />

      {/* Header */}
      <div className="relative z-10 mb-6 text-right">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-semibold mb-2.5">
          <HelpCircle className="w-3.5 h-3.5" />
          <span>راهنمای جامع شهروندی</span>
        </div>
        <h2 className="text-xl font-bold text-white mb-1">
          سوالات متداول و راهنمای ثبت گزارش
        </h2>
        <p className="text-xs text-slate-400">
          پاسخ به سوالات پرتکرار پیرامون نحوه ثبت، پیگیری و تضمین محرمانگی گزارش‌ها
        </p>
      </div>

      {/* Accordion List */}
      <div className="relative z-10 space-y-1">
        {faqData.map((item) => (
          <AccordionItem
            key={item.value}
            value={item.value}
            title={item.title}
            icon={item.icon}
            content={item.content}
            isOpen={openItem === item.value}
            onToggle={handleToggle}
          />
        ))}
      </div>
    </div>
  );
}

export default FAQAccordion;
