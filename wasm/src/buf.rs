use std::rc::Rc;
use std::cell::RefCell;
use std::ops::{ Deref, DerefMut };

pub struct BufPool {
    pool: Rc<RefCell<Vec<Vec<u8>>>>,
    capacity: usize,
}

impl BufPool {

    pub fn new(init_size: usize, capacity: usize) -> Self {
        BufPool {
            pool: Rc::new(RefCell::new(vec![Vec::with_capacity(capacity); init_size])),
            capacity,
        }
    }

    /// `pop` takes a `BufferPool` and returns a `Buffer`
	/// 
	/// The `BufferPool` is a `Rc<RefCell<Vec<Vec<u8>>>>` and the `Buffer` is a `Rc<RefCell<Vec<u8>>>`
	/// 
	/// The `BufferPool` is a `Vec` of `Vec`s of `u8`s. The `Buffer` is a `Vec` of `u8`s
	/// 
	/// Returns:
	/// 
	/// A Buffer struct.
	pub fn pop(&self) -> Buffer {
        let buf = self
            .pool
            .borrow_mut()
            .pop()
            .unwrap_or_else(|| Vec::with_capacity(self.capacity));
        Buffer {
            pool: self.pool.clone(),
            inner: buf,
        }
    }
}

pub struct Buffer {
    pool: Rc<RefCell<Vec<Vec<u8>>>>,
    inner: Vec<u8>,
}

impl Drop for Buffer {
    fn drop(&mut self) {
        self.pool
            .borrow_mut()
            .push(std::mem::take(&mut self.inner))
    }
}

impl Deref for Buffer {
    type Target = Vec<u8>;
    fn deref(&self) -> &Self::Target {
        &self.inner
    }
}

impl DerefMut for Buffer {
    fn deref_mut(&mut self) -> &mut Self::Target {
        &mut self.inner
    }
}
