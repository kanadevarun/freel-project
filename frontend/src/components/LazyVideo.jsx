import { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react';

/**
 * LazyVideo — defers src injection until the element is near the viewport.
 *
 * Props:
 *  src          – video URL (required)
 *  className    – CSS class(es) to apply to the <video> element
 *  poster       – optional poster image URL shown before playback
 *  rootMargin   – IntersectionObserver rootMargin (default '400px')
 *  eager        – if true, loads immediately (for above-the-fold hero videos)
 *  style        – optional inline styles
 *
 * Supports ref forwarding so parent components can access the <video> element.
 */
const LazyVideo = forwardRef(function LazyVideo(
  { src, className = '', poster, rootMargin = '400px', eager = false, style },
  ref
) {
  const [ready, setReady] = useState(eager);
  const videoRef = useRef(null);

  // Expose the underlying <video> DOM node via forwarded ref
  useImperativeHandle(ref, () => videoRef.current);

  useEffect(() => {
    if (eager) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setReady(true);
          observer.disconnect();
        }
      },
      { rootMargin }
    );

    const el = videoRef.current;
    if (el) observer.observe(el);
    return () => observer.disconnect();
  }, [eager, rootMargin]);

  // Once src is set, force play (autoPlay may not fire after src is injected dynamically)
  useEffect(() => {
    if (ready && videoRef.current) {
      const el = videoRef.current;
      el.load();
      el.play().catch(() => {
        // Autoplay blocked — browser will handle silently
      });
    }
  }, [ready]);

  return (
    <video
      ref={videoRef}
      className={className}
      muted
      loop
      playsInline
      preload={eager ? 'auto' : 'none'}
      poster={poster}
      style={style}
    >
      {ready && <source src={src} type="video/mp4" />}
    </video>
  );
});

export default LazyVideo;
