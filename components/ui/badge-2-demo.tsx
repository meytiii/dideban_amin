import { Badge, BadgeDot } from '@/components/ui/badge-2';
import { Clock, Key, Lock, ShieldCheck } from 'lucide-react';

export default function Component() {
  return (
    <div className="flex flex-col gap-6 p-6">
      <div className="flex flex-wrap items-center gap-3">
        <Badge appearance="outline" className="bg-[rgba(245,152,39,0.14)] border-[rgba(245,152,39,0.35)] text-[#F59827]">
          <BadgeDot />
          مدیر ارشد سامانه
        </Badge>
        <Badge variant="info" appearance="outline">
          <BadgeDot />
          اداره کل
        </Badge>
        <Badge variant="primary" appearance="outline">
          <BadgeDot />
          نواحی و مناطق
        </Badge>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <Badge variant="success" appearance="outline">
          <BadgeDot />
          فعال
        </Badge>
        <Badge variant="destructive" appearance="outline">
          <BadgeDot />
          مسدود / غیرفعال
        </Badge>
        <Badge variant="destructive" appearance="outline">
          <Clock className="size-3.5" />
          منقضی شده
        </Badge>
        <Badge variant="warning" appearance="outline">
          <Key className="size-3.5" />
          الزام تغییر رمز
        </Badge>
        <Badge variant="warning" appearance="outline">
          <Lock className="size-3.5" />
          قفل موقت
        </Badge>
        <Badge variant="success" appearance="outline">
          <ShieldCheck className="size-3.5" />
          ۲FA فعال
        </Badge>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <Badge>Primary</Badge>
        <Badge variant="secondary">Secondary</Badge>
        <Badge variant="outline">Outline</Badge>
        <Badge variant="destructive">Destructive</Badge>
        <Badge variant="success">Success</Badge>
        <Badge variant="warning">Warning</Badge>
        <Badge variant="info">Info</Badge>
      </div>
    </div>
  );
}
