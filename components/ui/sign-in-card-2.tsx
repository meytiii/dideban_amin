'use client'
import React, { useState } from 'react';
import { motion, AnimatePresence, useMotionValue, useTransform } from 'framer-motion';
import { User, Lock, Eye, EyeOff, ArrowLeft, ShieldCheck } from 'lucide-react';
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
  return (
    <input
      type={type}
      data-slot="input"
      className={cn(
        "file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground dark:bg-input/30 border-input flex h-10 w-full min-w-0 rounded-xl border bg-transparent px-3.5 py-2 text-base shadow-xs transition-[color,box-shadow,border-color] outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
        "focus-visible:border-blue-500 focus-visible:ring-blue-500/30 focus-visible:ring-[3px]",
        "aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive",
        className
      )}
      {...props}
    />
  )
}

export function SignInCard({
  title = "ورود امن - پنل حراست",
  subtitle = "اداره کل آموزش و پرورش استان همدان",
  onSubmit,
}: {
  title?: string;
  subtitle?: string;
  onSubmit?: (data: { username: string; password: string; rememberMe: boolean }) => void;
}) {
  const [showPassword, setShowPassword] = useState(false);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [focusedInput, setFocusedInput] = useState<string | null>(null);
  const [rememberMe, setRememberMe] = useState(false);

  // 3D card tilt effect
  const mouseX = useMotionValue(0);
  const mouseY = useMotionValue(0);
  const rotateX = useTransform(mouseY, [-300, 300], [10, -10]);
  const rotateY = useTransform(mouseX, [-300, 300], [-10, 10]);

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    mouseX.set(e.clientX - rect.left - rect.width / 2);
    mouseY.set(e.clientY - rect.top - rect.height / 2);
  };

  const handleMouseLeave = () => {
    mouseX.set(0);
    mouseY.set(0);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    onSubmit?.({ username, password, rememberMe });
    setTimeout(() => setIsLoading(false), 1500);
  };

  return (
    <div className="min-h-screen w-full bg-[#070b14] relative overflow-hidden flex items-center justify-center p-4">
      {/* Background ambient lighting */}
      <div className="absolute inset-0 bg-gradient-to-b from-blue-950/30 via-[#070b14]/70 to-[#020617]" />

      {/* Top radial glow */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[120vh] h-[60vh] rounded-b-[50%] bg-blue-500/10 blur-[90px] pointer-events-none" />
      <motion.div
        className="absolute top-0 left-1/2 -translate-x-1/2 w-[90vh] h-[50vh] rounded-b-full bg-blue-400/10 blur-[70px] pointer-events-none"
        animate={{
          opacity: [0.15, 0.35, 0.15],
          scale: [0.98, 1.03, 0.98]
        }}
        transition={{ duration: 7, repeat: Infinity, repeatType: "mirror" }}
      />

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-md relative z-10"
        style={{ perspective: 1500 }}
      >
        <motion.div
          className="relative"
          style={{ rotateX, rotateY }}
          onMouseMove={handleMouseMove}
          onMouseLeave={handleMouseLeave}
          whileHover={{ z: 10 }}
        >
          <div className="relative group">
            {/* Ambient Card glow */}
            <motion.div 
              className="absolute -inset-[1px] rounded-3xl opacity-0 group-hover:opacity-70 transition-opacity duration-700 pointer-events-none"
              animate={{
                boxShadow: [
                  "0 0 20px 2px rgba(59,130,246,0.1)",
                  "0 0 35px 8px rgba(59,130,246,0.22)",
                  "0 0 20px 2px rgba(59,130,246,0.1)"
                ],
                opacity: [0.3, 0.7, 0.3]
              }}
              transition={{ duration: 4, repeat: Infinity, ease: "easeInOut", repeatType: "mirror" }}
            />

            {/* Traveling light beams on borders */}
            <div className="absolute -inset-[1px] rounded-3xl overflow-hidden pointer-events-none">
              {/* Top beam */}
              <motion.div 
                className="absolute top-0 left-0 h-[3px] w-[50%] bg-gradient-to-r from-transparent via-blue-400 to-transparent opacity-85"
                animate={{ left: ["-50%", "100%"], opacity: [0.3, 0.9, 0.3] }}
                transition={{ left: { duration: 2.5, ease: "easeInOut", repeat: Infinity, repeatDelay: 0.8 }, opacity: { duration: 1.2, repeat: Infinity, repeatType: "mirror" } }}
              />
              {/* Right beam */}
              <motion.div 
                className="absolute top-0 right-0 h-[50%] w-[3px] bg-gradient-to-b from-transparent via-blue-400 to-transparent opacity-85"
                animate={{ top: ["-50%", "100%"], opacity: [0.3, 0.9, 0.3] }}
                transition={{ top: { duration: 2.5, ease: "easeInOut", repeat: Infinity, repeatDelay: 0.8, delay: 0.6 }, opacity: { duration: 1.2, repeat: Infinity, repeatType: "mirror", delay: 0.6 } }}
              />
              {/* Bottom beam */}
              <motion.div 
                className="absolute bottom-0 right-0 h-[3px] w-[50%] bg-gradient-to-r from-transparent via-blue-400 to-transparent opacity-85"
                animate={{ right: ["-50%", "100%"], opacity: [0.3, 0.9, 0.3] }}
                transition={{ right: { duration: 2.5, ease: "easeInOut", repeat: Infinity, repeatDelay: 0.8, delay: 1.2 }, opacity: { duration: 1.2, repeat: Infinity, repeatType: "mirror", delay: 1.2 } }}
              />
              {/* Left beam */}
              <motion.div 
                className="absolute bottom-0 left-0 h-[50%] w-[3px] bg-gradient-to-b from-transparent via-blue-400 to-transparent opacity-85"
                animate={{ bottom: ["-50%", "100%"], opacity: [0.3, 0.9, 0.3] }}
                transition={{ bottom: { duration: 2.5, ease: "easeInOut", repeat: Infinity, repeatDelay: 0.8, delay: 1.8 }, opacity: { duration: 1.2, repeat: Infinity, repeatType: "mirror", delay: 1.8 } }}
              />
            </div>

            {/* Glass Card Surface */}
            <div className="relative bg-[#0f172a]/90 backdrop-blur-2xl rounded-3xl p-8 sm:p-10 border border-blue-500/20 shadow-2xl overflow-hidden">
              {/* Logo / Crest Header */}
              <div className="text-center space-y-2 mb-7">
                <motion.div
                  initial={{ scale: 0.7, opacity: 0 }}
                  animate={{ scale: 1, opacity: 1 }}
                  transition={{ type: "spring", duration: 0.8 }}
                  className="mx-auto w-16 h-16 rounded-2xl border border-blue-400/40 bg-blue-500/15 flex items-center justify-center relative overflow-hidden text-blue-400 shadow-inner"
                >
                  <ShieldCheck className="w-9 h-9 text-blue-400 drop-shadow-[0_0_12px_rgba(59,130,246,0.6)]" />
                </motion.div>

                <motion.h1
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.2 }}
                  className="text-xl font-bold text-white tracking-tight"
                >
                  {title}
                </motion.h1>

                <motion.p
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  transition={{ delay: 0.3 }}
                  className="text-slate-400 text-xs sm:text-sm"
                >
                  {subtitle}
                </motion.p>
              </div>

              {/* Login Form */}
              <form onSubmit={handleSubmit} className="space-y-5">
                <div className="space-y-4">
                  {/* Username Field */}
                  <div className="space-y-1.5 text-right">
                    <label className="text-sm font-semibold text-slate-200">نام کاربری</label>
                    <div className="relative flex items-center">
                      <User className={`absolute right-4 w-5 h-5 transition-colors duration-200 ${
                        focusedInput === "username" ? 'text-blue-400' : 'text-slate-400'
                      }`} />
                      <Input
                        type="text"
                        placeholder="نام کاربری اداری"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        onFocus={() => setFocusedInput("username")}
                        onBlur={() => setFocusedInput(null)}
                        required
                        className="w-full h-12 bg-slate-900/80 border-slate-700/80 text-white placeholder:text-slate-500 pr-12 pl-4 text-right rounded-xl text-base"
                      />
                    </div>
                  </div>

                  {/* Password Field */}
                  <div className="space-y-1.5 text-right">
                    <label className="text-sm font-semibold text-slate-200">رمز عبور</label>
                    <div className="relative flex items-center">
                      <Lock className={`absolute right-4 w-5 h-5 transition-colors duration-200 ${
                        focusedInput === "password" ? 'text-blue-400' : 'text-slate-400'
                      }`} />
                      <Input
                        type={showPassword ? "text" : "password"}
                        placeholder="رمز عبور"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        onFocus={() => setFocusedInput("password")}
                        onBlur={() => setFocusedInput(null)}
                        required
                        className="w-full h-12 bg-slate-900/80 border-slate-700/80 text-white placeholder:text-slate-500 pr-12 pl-12 text-right rounded-xl text-base"
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="absolute left-3.5 text-slate-400 hover:text-white transition-colors p-1.5 rounded-lg hover:bg-white/5"
                        aria-label="تغییر نمایش رمز"
                      >
                        {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                      </button>
                    </div>
                  </div>
                </div>

                {/* Remember Me */}
                <div className="flex items-center justify-start gap-2 pt-1 text-right">
                  <input
                    id="remember-me"
                    type="checkbox"
                    checked={rememberMe}
                    onChange={(e) => setRememberMe(e.target.checked)}
                    className="w-4 h-4 rounded border-slate-700 bg-slate-900/80 text-blue-600 focus:ring-blue-500/30"
                  />
                  <label htmlFor="remember-me" className="text-xs text-slate-400 cursor-pointer select-none">
                    مرا به خاطر بسپار
                  </label>
                </div>

                {/* Submit Button */}
                <motion.button
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  type="submit"
                  disabled={isLoading}
                  className="w-full relative group/btn mt-5 cursor-pointer"
                >
                  <div className="absolute inset-0 bg-blue-600/30 rounded-xl blur-md opacity-0 group-hover/btn:opacity-100 transition-opacity duration-300" />
                  <div className="relative overflow-hidden bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-500 hover:to-blue-600 text-white font-bold h-11 rounded-xl shadow-lg transition-all duration-300 flex items-center justify-center gap-2">
                    <AnimatePresence mode="wait">
                      {isLoading ? (
                        <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                      ) : (
                        <span className="flex items-center gap-1.5 text-sm">
                          <span>ورود به سیستم</span>
                          <ArrowLeft className="w-4 h-4 group-hover/btn:-translate-x-1 transition-transform duration-300" />
                        </span>
                      )}
                    </AnimatePresence>
                  </div>
                </motion.button>
              </form>
            </div>
          </div>
        </motion.div>
      </motion.div>
    </div>
  );
}

export default SignInCard;
