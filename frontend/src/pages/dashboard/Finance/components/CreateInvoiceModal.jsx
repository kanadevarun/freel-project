import React, { useState, useEffect, useMemo } from 'react';
import {
  X,
  Plus,
  Trash2,
  FileText,
  Ship,
  Bookmark,
  DollarSign,
  Calendar,
  Layers,
  AlertCircle,
  CheckCircle2,
  User,
  ArrowRight
} from 'lucide-react';
import api from '../../../../services/api';
import './CreateInvoiceModal.css';

export default function CreateInvoiceModal({ isOpen, onClose, onSuccess, showToast, editingInvoice = null }) {
  if (!isOpen) return null;

  const isEditMode = Boolean(editingInvoice);

  // Source selection state: 'MANUAL' | 'SHIPMENT' | 'BOOKING' | 'QUOTATION'
  const [sourceType, setSourceType] = useState('MANUAL');

  // Tenant reference lists loaded from backend
  const [customers, setCustomers] = useState([]);
  const [shipments, setShipments] = useState([]);
  const [bookings, setBookings] = useState([]);
  const [quotations, setQuotations] = useState([]);
  const [loadingRefs, setLoadingRefs] = useState(true);

  // Form input state
  const [selectedCustomerId, setSelectedCustomerId] = useState('');
  const [customerName, setCustomerName] = useState('');
  const [customerCountry, setCustomerCountry] = useState('USA');
  const [currency, setCurrency] = useState('USD');

  // Dates
  const todayStr = new Date().toISOString().split('T')[0];
  const defaultDueStr = new Date(Date.now() + 15 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
  const [invoiceDate, setInvoiceDate] = useState(todayStr);
  const [dueDate, setDueDate] = useState(defaultDueStr);

  // Linked Source IDs & Labels
  const [selectedShipmentId, setSelectedShipmentId] = useState('');
  const [shipmentNumber, setShipmentNumber] = useState('');
  const [selectedBookingId, setSelectedBookingId] = useState('');
  const [bookingNumber, setBookingNumber] = useState('');
  const [selectedQuotationId, setSelectedQuotationId] = useState('');
  const [quoteNumber, setQuoteNumber] = useState('');

  // Route context
  const [route, setRoute] = useState('');
  const [origin, setOrigin] = useState('');
  const [destination, setDestination] = useState('');

  // Financial Line Items
  const [lineItems, setLineItems] = useState([
    {
      id: 1,
      description: 'Ocean Freight (40ft FCL High Cube Container)',
      serviceCategory: 'Freight',
      quantity: 1,
      unitPrice: 2450.00
    }
  ]);

  // Adjustments
  const [taxAmount, setTaxAmount] = useState(0);
  const [discountAmount, setDiscountAmount] = useState(0);

  // Status & Submission
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [formError, setFormError] = useState(null);

  // Load tenant datasets on mount
  useEffect(() => {
    let isMounted = true;
    async function loadTenantData() {
      try {
        setLoadingRefs(true);

        // 1. Fetch Customers
        const custRes = await api.get('/api/v1/customers');
        const custData = custRes?.customers || custRes?.items || (Array.isArray(custRes) ? custRes : []);
        if (isMounted) setCustomers(custData);

        // 2. Fetch Shipments
        const shipRes = await api.get('/api/v1/shipments?workspace=true&page=1&limit=50');
        const shipData = shipRes?.shipments || shipRes?.items || (Array.isArray(shipRes) ? shipRes : []);
        if (isMounted) setShipments(shipData);

        // 3. Fetch Bookings
        const bookRes = await api.get('/api/v1/bookings');
        const bookData = bookRes?.items || bookRes?.bookings || (Array.isArray(bookRes) ? bookRes : []);
        if (isMounted) setBookings(bookData);

        // 4. Fetch Quotations
        const quoteRes = await api.get('/api/v1/quotations');
        const quoteData = quoteRes?.quotations || quoteRes?.items || (Array.isArray(quoteRes) ? quoteRes : []);
        if (isMounted) setQuotations(quoteData);

        // Auto select or prefill from editingInvoice
        if (editingInvoice) {
          setSelectedCustomerId(editingInvoice.customer_id || editingInvoice.customerId || '');
          setCustomerName(editingInvoice.customer_name || editingInvoice.customer || '');
          setCustomerCountry(editingInvoice.customer_country || editingInvoice.customerCountry || 'USA');
          setCurrency(editingInvoice.currency || 'USD');
          if (editingInvoice.invoice_date || editingInvoice.invoiceDate) {
            const d = new Date(editingInvoice.invoice_date || editingInvoice.invoiceDate).toISOString().split('T')[0];
            setInvoiceDate(d);
          }
          if (editingInvoice.due_date || editingInvoice.dueDate) {
            const d = new Date(editingInvoice.due_date || editingInvoice.dueDate).toISOString().split('T')[0];
            setDueDate(d);
          }
          setRoute(editingInvoice.route || '');
          setTaxAmount(editingInvoice.tax_amount || editingInvoice.tax || 0);
          setDiscountAmount(editingInvoice.discount_amount || editingInvoice.discount || 0);

          if (editingInvoice.line_items || editingInvoice.lineItems) {
            const items = (editingInvoice.line_items || editingInvoice.lineItems).map((it, idx) => ({
              id: it.id || idx + 1,
              description: it.description,
              serviceCategory: it.service_category || it.serviceCategory || 'Freight',
              quantity: Number(it.quantity || 1),
              unitPrice: Number(it.unit_price || it.rate || 0)
            }));
            if (items.length > 0) setLineItems(items);
          }
        } else if (custData.length > 0) {
          setSelectedCustomerId(custData[0].id);
          setCustomerName(custData[0].name);
          setCustomerCountry(custData[0].country || 'USA');
        }
      } catch (err) {
        console.warn('Failed to load reference datasets for invoice creation:', err);
      } finally {
        if (isMounted) setLoadingRefs(false);
      }
    }

    loadTenantData();
    return () => { isMounted = false; };
  }, []);

  // Handle Customer Selection
  const handleCustomerChange = (e) => {
    const custId = e.target.value;
    setSelectedCustomerId(custId);
    const found = customers.find((c) => String(c.id) === String(custId));
    if (found) {
      setCustomerName(found.name);
      setCustomerCountry(found.country || 'USA');
    }
  };

  // Handle Source Type Switching
  const handleSourceTypeChange = (type) => {
    setSourceType(type);
    setFormError(null);

    // Reset linked values
    setSelectedShipmentId('');
    setShipmentNumber('');
    setSelectedBookingId('');
    setBookingNumber('');
    setSelectedQuotationId('');
    setQuoteNumber('');
  };

  // Handle Shipment Selection
  const handleShipmentSelect = (shipIdStr) => {
    setSelectedShipmentId(shipIdStr);
    const found = shipments.find((s) => String(s.id) === String(shipIdStr));
    if (found) {
      const num = found.shipment_number || found.rfq_number || `SH-2026-${found.id}`;
      setShipmentNumber(num);
      const routeStr = found.route || (found.origin_port && found.destination_port ? `${found.origin_port} ➔ ${found.destination_port}` : 'Shanghai ➔ Los Angeles');
      setRoute(routeStr);
      setOrigin(found.origin_port || '');
      setDestination(found.destination_port || '');

      if (found.customer_name) {
        setCustomerName(found.customer_name);
        const matchCust = customers.find((c) => c.name.toLowerCase().includes(found.customer_name.toLowerCase()));
        if (matchCust) setSelectedCustomerId(matchCust.id);
      }

      if (found.booking_number) {
        setBookingNumber(found.booking_number);
        if (found.booking_id) setSelectedBookingId(found.booking_id);
      }

      setLineItems([
        {
          id: Date.now(),
          description: `Ocean Freight Charges (${num})`,
          serviceCategory: 'Freight',
          quantity: 1,
          unitPrice: 3250.00
        },
        {
          id: Date.now() + 1,
          description: 'Port Handling & Documentation Fee',
          serviceCategory: 'Documentation',
          quantity: 1,
          unitPrice: 350.00
        }
      ]);
    }
  };

  // Handle Booking Selection
  const handleBookingSelect = (bookIdStr) => {
    setSelectedBookingId(bookIdStr);
    const found = bookings.find((b) => String(b.id) === String(bookIdStr));
    if (found) {
      const num = found.booking_number || `BK-${found.id}`;
      setBookingNumber(num);
      const routeStr = found.origin_port && found.destination_port ? `${found.origin_port} ➔ ${found.destination_port}` : 'Nhava Sheva ➔ Rotterdam';
      setRoute(routeStr);
      setOrigin(found.origin_port || '');
      setDestination(found.destination_port || '');

      if (found.customer_name) {
        setCustomerName(found.customer_name);
        const matchCust = customers.find((c) => c.name.toLowerCase().includes(found.customer_name.toLowerCase()));
        if (matchCust) setSelectedCustomerId(matchCust.id);
      }

      if (found.shipment_id) {
        setSelectedShipmentId(found.shipment_id);
        setShipmentNumber(`SH-2026-${found.shipment_id}`);
      }

      const sellPrice = Number(found.quote_sell_price || 4800.00);
      setLineItems([
        {
          id: Date.now(),
          description: `Freight Booking ${num} (${found.cargo_summary || 'FCL Cargo'})`,
          serviceCategory: 'Freight',
          quantity: 1,
          unitPrice: sellPrice
        }
      ]);
    }
  };

  // Handle Quotation Selection
  const handleQuotationSelect = (quoteIdStr) => {
    setSelectedQuotationId(quoteIdStr);
    const found = quotations.find((q) => String(q.id) === String(quoteIdStr));
    if (found) {
      const num = found.quotation_number || `QT-2026-${found.id}`;
      setQuoteNumber(num);
      const routeStr = found.route || (found.origin && found.destination ? `${found.origin} ➔ ${found.destination}` : 'New York ➔ Nhava Sheva');
      setRoute(routeStr);
      setOrigin(found.origin || '');
      setDestination(found.destination || '');

      if (found.customer_name) {
        setCustomerName(found.customer_name);
        const matchCust = customers.find((c) => c.name.toLowerCase().includes(found.customer_name.toLowerCase()));
        if (matchCust) setSelectedCustomerId(matchCust.id);
      }

      const amt = Number(found.total_amount || 2950.00);
      setLineItems([
        {
          id: Date.now(),
          description: `Charges as per Quotation ${num}`,
          serviceCategory: 'Freight',
          quantity: 1,
          unitPrice: amt
        }
      ]);
    }
  };

  // Line item manipulation
  const handleAddLineItem = () => {
    setLineItems((prev) => [
      ...prev,
      {
        id: Date.now(),
        description: '',
        serviceCategory: 'Freight',
        quantity: 1,
        unitPrice: 0.00
      }
    ]);
  };

  const handleUpdateLineItem = (id, field, value) => {
    setLineItems((prev) =>
      prev.map((item) => {
        if (item.id === id) {
          const updated = { ...item, [field]: value };
          return updated;
        }
        return item;
      })
    );
  };

  const handleRemoveLineItem = (id) => {
    if (lineItems.length <= 1) {
      showToast && showToast('An invoice must contain at least 1 line item', 'warning');
      return;
    }
    setLineItems((prev) => prev.filter((item) => item.id !== id));
  };

  // Live Financial Preview Calculations
  const calculatedSubtotal = useMemo(() => {
    return lineItems.reduce((acc, item) => {
      const qty = Number(item.quantity || 0);
      const price = Number(item.unitPrice || 0);
      return acc + qty * price;
    }, 0);
  }, [lineItems]);

  const calculatedTotal = useMemo(() => {
    const sub = calculatedSubtotal;
    const tax = Number(taxAmount || 0);
    const disc = Number(discountAmount || 0);
    const tot = sub + tax - disc;
    return tot > 0 ? tot : 0;
  }, [calculatedSubtotal, taxAmount, discountAmount]);

  // Submit Handler
  const handleSubmit = async (targetStatus) => {
    try {
      setFormError(null);

      if (!customerName.trim() && !selectedCustomerId) {
        setFormError('Please select or specify a Customer.');
        return;
      }

      if (lineItems.length === 0) {
        setFormError('Please add at least one valid line item.');
        return;
      }

      for (let i = 0; i < lineItems.length; i++) {
        if (!lineItems[i].description.trim()) {
          setFormError(`Line item #${i + 1} requires a description.`);
          return;
        }
      }

      setIsSubmitting(true);

      const payload = {
        customer_id: selectedCustomerId ? Number(selectedCustomerId) : 1,
        customer_name: customerName || 'Global Traders Inc.',
        customer_country: customerCountry || 'USA',
        shipment_id: selectedShipmentId ? Number(selectedShipmentId) : null,
        shipment_number: shipmentNumber,
        booking_id: selectedBookingId ? Number(selectedBookingId) : null,
        booking_number: bookingNumber,
        quotation_id: selectedQuotationId ? Number(selectedQuotationId) : null,
        quote_number: quoteNumber,
        route: route || 'Shanghai ➔ Los Angeles',
        origin: origin,
        destination: destination,
        invoice_date: invoiceDate,
        due_date: dueDate,
        currency: currency,
        subtotal: calculatedSubtotal,
        tax_amount: Number(taxAmount || 0),
        discount_amount: Number(discountAmount || 0),
        status: targetStatus, // 'Draft' or 'Issued'
        line_items: lineItems.map((item) => ({
          description: item.description,
          service_category: item.serviceCategory || 'Freight',
          quantity: Number(item.quantity || 1),
          unit_price: Number(item.unitPrice || 0)
        }))
      };

      let resultInvoice;
      if (isEditMode && editingInvoice?.id) {
        resultInvoice = await api.put(`/api/v1/invoices/${editingInvoice.id}`, payload);
      } else {
        resultInvoice = await api.post('/api/v1/invoices', payload);
      }

      if (showToast) {
        showToast(
          isEditMode
            ? `Invoice ${resultInvoice?.invoice_number || ''} updated successfully`
            : `Invoice ${resultInvoice?.invoice_number || ''} created successfully (${targetStatus})`,
          'success'
        );
      }

      if (onSuccess) {
        onSuccess(resultInvoice);
      }

      onClose();
    } catch (err) {
      console.error('Failed to create invoice:', err);
      setFormError(err.message || 'Failed to create invoice in backend.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="create-invoice-modal-overlay">
      <div className="create-invoice-modal-container">
        
        {/* Modal Header */}
        <div className="create-invoice-header">
          <div className="header-icon-box">
            <Plus size={20} className="icon-blue" />
          </div>
          <div className="header-text">
            <h2>{isEditMode ? `Edit Draft Invoice (${editingInvoice?.invoiceNumber || editingInvoice?.invoice_number || ''})` : 'Create New Customer Invoice'}</h2>
            <p>{isEditMode ? 'Modify draft invoice line items, dates, and charges' : 'Generate a persistent freight invoice for your customer'}</p>
          </div>
          <button className="close-modal-btn" onClick={onClose} disabled={isSubmitting}>
            <X size={18} />
          </button>
        </div>

        {/* Modal Body Scrollable Content */}
        <div className="create-invoice-body">
          
          {/* Form Error Alert */}
          {formError && (
            <div className="create-invoice-alert-banner">
              <AlertCircle size={16} />
              <span>{formError}</span>
            </div>
          )}

          {/* 1. CREATION SOURCE SELECTION */}
          <div className="form-section-card">
            <h3 className="section-title">
              <Layers size={15} /> 1. Select Invoice Creation Source
            </h3>

            <div className="source-type-grid">
              <button
                type="button"
                className={`source-card-btn ${sourceType === 'MANUAL' ? 'selected' : ''}`}
                onClick={() => handleSourceTypeChange('MANUAL')}
              >
                <div className="source-card-header">
                  <FileText size={18} />
                  <span>Manual Invoice</span>
                </div>
                <p>Create a standalone customer invoice from scratch.</p>
              </button>

              <button
                type="button"
                className={`source-card-btn ${sourceType === 'SHIPMENT' ? 'selected' : ''}`}
                onClick={() => handleSourceTypeChange('SHIPMENT')}
              >
                <div className="source-card-header">
                  <Ship size={18} />
                  <span>From Shipment</span>
                </div>
                <p>Prefill customer, route, and charges from an active shipment.</p>
              </button>

              <button
                type="button"
                className={`source-card-btn ${sourceType === 'BOOKING' ? 'selected' : ''}`}
                onClick={() => handleSourceTypeChange('BOOKING')}
              >
                <div className="source-card-header">
                  <Bookmark size={18} />
                  <span>From Booking</span>
                </div>
                <p>Link invoice directly to a confirmed freight booking.</p>
              </button>

              <button
                type="button"
                className={`source-card-btn ${sourceType === 'QUOTATION' ? 'selected' : ''}`}
                onClick={() => handleSourceTypeChange('QUOTATION')}
              >
                <div className="source-card-header">
                  <DollarSign size={18} />
                  <span>From Quotation</span>
                </div>
                <p>Copy approved rate quotes and charges into an invoice.</p>
              </button>
            </div>

            {/* Dynamic Source Selection Dropdowns */}
            {sourceType === 'SHIPMENT' && (
              <div className="source-selector-box">
                <label>Select Active Shipment *</label>
                <select
                  value={selectedShipmentId}
                  onChange={(e) => handleShipmentSelect(e.target.value)}
                  className="form-control-select"
                >
                  <option value="">-- Choose Shipment --</option>
                  {shipments.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.shipment_number || s.rfq_number || `SH-2026-${s.id}`} — {s.customer_name || 'Global Freight Customer'} ({s.origin_port || 'INNSA'} ➔ {s.destination_port || 'NLRTM'})
                    </option>
                  ))}
                </select>
              </div>
            )}

            {sourceType === 'BOOKING' && (
              <div className="source-selector-box">
                <label>Select Confirmed Freight Booking *</label>
                <select
                  value={selectedBookingId}
                  onChange={(e) => handleBookingSelect(e.target.value)}
                  className="form-control-select"
                >
                  <option value="">-- Choose Booking --</option>
                  {bookings.map((b) => (
                    <option key={b.id} value={b.id}>
                      {b.booking_number || `BK-${b.id}`} — {b.customer_name || 'Customer'} ({b.origin_port || 'INNSA'} ➔ {b.destination_port || 'NLRTM'})
                    </option>
                  ))}
                </select>
              </div>
            )}

            {sourceType === 'QUOTATION' && (
              <div className="source-selector-box">
                <label>Select Eligible Quotation *</label>
                <select
                  value={selectedQuotationId}
                  onChange={(e) => handleQuotationSelect(e.target.value)}
                  className="form-control-select"
                >
                  <option value="">-- Choose Quotation --</option>
                  {quotations.map((q) => (
                    <option key={q.id} value={q.id}>
                      {q.quotation_number || `QT-2026-${q.id}`} — {q.customer_name || 'Client'} (${q.total_amount || 2950} USD)
                    </option>
                  ))}
                </select>
              </div>
            )}
          </div>

          {/* 2. INVOICE HEADER & CUSTOMER DETAILS */}
          <div className="form-section-card">
            <h3 className="section-title">
              <User size={15} /> 2. Invoice Details & Customer Information
            </h3>

            <div className="form-grid-2col">
              <div className="form-group">
                <label>Bill To Customer *</label>
                <select
                  value={selectedCustomerId}
                  onChange={handleCustomerChange}
                  className="form-control-select"
                >
                  <option value="">-- Select Customer --</option>
                  {customers.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name} ({c.customer_code || 'CUST'})
                    </option>
                  ))}
                </select>
              </div>

              <div className="form-group">
                <label>Customer Name / Company Label</label>
                <input
                  type="text"
                  value={customerName}
                  onChange={(e) => setCustomerName(e.target.value)}
                  placeholder="e.g. Global Traders Inc."
                  className="form-control-input"
                />
              </div>

              <div className="form-group">
                <label>Invoice Date *</label>
                <input
                  type="date"
                  value={invoiceDate}
                  onChange={(e) => setInvoiceDate(e.target.value)}
                  className="form-control-input"
                />
              </div>

              <div className="form-group">
                <label>Payment Due Date *</label>
                <input
                  type="date"
                  value={dueDate}
                  onChange={(e) => setDueDate(e.target.value)}
                  className="form-control-input"
                />
              </div>

              <div className="form-group">
                <label>Currency *</label>
                <select
                  value={currency}
                  onChange={(e) => setCurrency(e.target.value)}
                  className="form-control-select"
                >
                  <option value="USD">USD ($ - US Dollar)</option>
                  <option value="EUR">EUR (€ - Euro)</option>
                  <option value="GBP">GBP (£ - British Pound)</option>
                  <option value="INR">INR (₹ - Indian Rupee)</option>
                  <option value="SGD">SGD (S$ - Singapore Dollar)</option>
                  <option value="CAD">CAD (C$ - Canadian Dollar)</option>
                </select>
              </div>

              <div className="form-group">
                <label>Route / Trade Lane</label>
                <input
                  type="text"
                  value={route}
                  onChange={(e) => setRoute(e.target.value)}
                  placeholder="e.g. Shanghai ➔ Los Angeles"
                  className="form-control-input"
                />
              </div>
            </div>
          </div>

          {/* 3. LINE ITEMS */}
          <div className="form-section-card">
            <div className="section-header-flex">
              <h3 className="section-title">
                <FileText size={15} /> 3. Invoice Charges & Line Items
              </h3>
              <button
                type="button"
                className="btn-add-item"
                onClick={handleAddLineItem}
              >
                <Plus size={14} /> Add Line Item
              </button>
            </div>

            <div className="line-items-table-wrapper">
              <table className="line-items-table">
                <thead>
                  <tr>
                    <th style={{ width: '40%' }}>Description *</th>
                    <th style={{ width: '20%' }}>Category</th>
                    <th style={{ width: '12%', textAlign: 'right' }}>Qty</th>
                    <th style={{ width: '18%', textAlign: 'right' }}>Rate ({currency})</th>
                    <th style={{ width: '10%', textAlign: 'center' }}></th>
                  </tr>
                </thead>
                <tbody>
                  {lineItems.map((item) => (
                    <tr key={item.id}>
                      <td>
                        <input
                          type="text"
                          value={item.description}
                          onChange={(e) => handleUpdateLineItem(item.id, 'description', e.target.value)}
                          placeholder="Charge description e.g. Ocean Freight 40ft Container"
                          className="table-input"
                        />
                      </td>
                      <td>
                        <select
                          value={item.serviceCategory}
                          onChange={(e) => handleUpdateLineItem(item.id, 'serviceCategory', e.target.value)}
                          className="table-select"
                        >
                          <option value="Freight">Freight</option>
                          <option value="Documentation">Documentation</option>
                          <option value="Customs">Customs</option>
                          <option value="Handling">Handling</option>
                          <option value="Terminal">Terminal</option>
                          <option value="Storage">Storage</option>
                          <option value="Other">Other</option>
                        </select>
                      </td>
                      <td>
                        <input
                          type="number"
                          min="1"
                          step="1"
                          value={item.quantity}
                          onChange={(e) => handleUpdateLineItem(item.id, 'quantity', e.target.value)}
                          className="table-input text-right"
                        />
                      </td>
                      <td>
                        <input
                          type="number"
                          min="0"
                          step="0.01"
                          value={item.unitPrice}
                          onChange={(e) => handleUpdateLineItem(item.id, 'unitPrice', e.target.value)}
                          className="table-input text-right"
                        />
                      </td>
                      <td style={{ textAlign: 'center' }}>
                        <button
                          type="button"
                          className="btn-remove-row"
                          onClick={() => handleRemoveLineItem(item.id)}
                          title="Remove item"
                        >
                          <Trash2 size={15} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* 4. TOTALS & SUMMARY */}
          <div className="form-section-card totals-card">
            <div className="totals-flex-container">
              <div className="totals-notes-left">
                <label>Internal Reference / Notes</label>
                <textarea
                  placeholder="Optional billing reference notes or payment instructions..."
                  className="notes-textarea"
                  rows={3}
                />
              </div>

              <div className="totals-summary-right">
                <div className="totals-row">
                  <span>Subtotal:</span>
                  <span className="totals-val">${calculatedSubtotal.toFixed(2)} {currency}</span>
                </div>

                <div className="totals-row input-row">
                  <span>Tax Amount:</span>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    value={taxAmount}
                    onChange={(e) => setTaxAmount(e.target.value)}
                    className="totals-input"
                  />
                </div>

                <div className="totals-row input-row">
                  <span>Discount Amount:</span>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    value={discountAmount}
                    onChange={(e) => setDiscountAmount(e.target.value)}
                    className="totals-input"
                  />
                </div>

                <div className="totals-divider" />

                <div className="totals-row grand-total-row">
                  <span>Total Amount Due:</span>
                  <span className="grand-total-val">${calculatedTotal.toFixed(2)} {currency}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Modal Footer Actions */}
        <div className="create-invoice-footer">
          <button
            type="button"
            className="btn-cancel-modal"
            onClick={onClose}
            disabled={isSubmitting}
          >
            Cancel
          </button>

          <div className="footer-submit-group">
            <button
              type="button"
              className="btn-draft-submit"
              onClick={() => handleSubmit('Draft')}
              disabled={isSubmitting}
            >
              Save as Draft
            </button>

            <button
              type="button"
              className="btn-primary-submit"
              onClick={() => handleSubmit('Issued')}
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Creating Invoice...' : 'Issue Invoice'}
            </button>
          </div>
        </div>

      </div>
    </div>
  );
}
