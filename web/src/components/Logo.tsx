interface LogoProps {
  size?: number;
  showText?: boolean;
}

/**
 * ReviewBuddy logo —— 评审闭环：开口圆环 + 对勾
 * 圆环代表模板/规则演进闭环，对勾代表评审通过与质量提升。
 * 圆环底色用主题主色 var(--color-primary)，跟随主题切换。
 */
export default function Logo({ size = 32, showText = true }: LogoProps) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
      <svg width={size} height={size} viewBox="0 0 48 48" fill="none" aria-label="ReviewBuddy">
        <rect width="48" height="48" rx="11" fill="var(--color-primary)" />
        <path
          d="M24 9 A15 15 0 1 1 11.8 15.2"
          stroke="#ffffff"
          strokeWidth="3.4"
          strokeLinecap="round"
          fill="none"
        />
        <path
          d="M16.5 23.5 L22 29 L33 16.5"
          stroke="#ffffff"
          strokeWidth="3.4"
          strokeLinecap="round"
          strokeLinejoin="round"
          fill="none"
        />
      </svg>
      {showText && (
        <span style={{ fontSize: size * 0.5, fontWeight: 600, color: 'var(--text-primary)' }}>
          ReviewBuddy
        </span>
      )}
    </div>
  );
}
