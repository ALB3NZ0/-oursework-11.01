--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.4

-- Started on 2025-11-10 16:43:45

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

ALTER TABLE ONLY public.reviews DROP CONSTRAINT reviews_iduser_fkey;
ALTER TABLE ONLY public.reviews DROP CONSTRAINT reviews_idproduct_fkey;
ALTER TABLE ONLY public.reports DROP CONSTRAINT reports_iduser_fkey;
ALTER TABLE ONLY public.productsizes DROP CONSTRAINT productsizes_idproduct_fkey;
ALTER TABLE ONLY public.products DROP CONSTRAINT products_idcategory_fkey;
ALTER TABLE ONLY public.products DROP CONSTRAINT products_idbrand_fkey;
ALTER TABLE ONLY public.orders DROP CONSTRAINT orders_iduser_fkey;
ALTER TABLE ONLY public.orderproducts DROP CONSTRAINT orderproducts_idproductsize_fkey;
ALTER TABLE ONLY public.orderproducts DROP CONSTRAINT orderproducts_idorder_fkey;
ALTER TABLE ONLY public.logs DROP CONSTRAINT logs_iduser_fkey;
ALTER TABLE ONLY public.favorites DROP CONSTRAINT favorites_iduser_fkey;
ALTER TABLE ONLY public.favorites DROP CONSTRAINT favorites_idproductsize_fkey;
ALTER TABLE ONLY public.basket DROP CONSTRAINT basket_iduser_fkey;
ALTER TABLE ONLY public.basket DROP CONSTRAINT basket_idproductsize_fkey;
DROP TRIGGER trg_decrease_stock ON public.orderproducts;
DROP TRIGGER trg_check_product_quantity ON public.productsizes;
DROP TRIGGER trg_add_default_sizes ON public.products;
ALTER TABLE ONLY public.users DROP CONSTRAINT users_pkey;
ALTER TABLE ONLY public.users DROP CONSTRAINT users_email_key;
ALTER TABLE ONLY public.reviews DROP CONSTRAINT reviews_pkey;
ALTER TABLE ONLY public.reports DROP CONSTRAINT reports_pkey;
ALTER TABLE ONLY public.productsizes DROP CONSTRAINT productsizes_pkey;
ALTER TABLE ONLY public.products DROP CONSTRAINT products_pkey;
ALTER TABLE ONLY public.orders DROP CONSTRAINT orders_pkey;
ALTER TABLE ONLY public.orderproducts DROP CONSTRAINT orderproducts_pkey;
ALTER TABLE ONLY public.logs DROP CONSTRAINT logs_pkey;
ALTER TABLE ONLY public.favorites DROP CONSTRAINT favorites_pkey;
ALTER TABLE ONLY public.categories DROP CONSTRAINT categories_pkey;
ALTER TABLE ONLY public.brands DROP CONSTRAINT brands_pkey;
ALTER TABLE ONLY public.basket DROP CONSTRAINT basket_pkey;
ALTER TABLE public.users ALTER COLUMN iduser DROP DEFAULT;
ALTER TABLE public.reviews ALTER COLUMN idreview DROP DEFAULT;
ALTER TABLE public.reports ALTER COLUMN idreport DROP DEFAULT;
ALTER TABLE public.productsizes ALTER COLUMN idproductsize DROP DEFAULT;
ALTER TABLE public.products ALTER COLUMN idproduct DROP DEFAULT;
ALTER TABLE public.orders ALTER COLUMN idorder DROP DEFAULT;
ALTER TABLE public.orderproducts ALTER COLUMN idorderproduct DROP DEFAULT;
ALTER TABLE public.logs ALTER COLUMN idlog DROP DEFAULT;
ALTER TABLE public.favorites ALTER COLUMN idfavorites DROP DEFAULT;
ALTER TABLE public.categories ALTER COLUMN idcategory DROP DEFAULT;
ALTER TABLE public.brands ALTER COLUMN idbrand DROP DEFAULT;
ALTER TABLE public.basket ALTER COLUMN idbasket DROP DEFAULT;
DROP SEQUENCE public.users_iduser_seq;
DROP VIEW public.topproducts;
DROP VIEW public.topcustomers;
DROP TABLE public.users;
DROP SEQUENCE public.reviews_idreview_seq;
DROP SEQUENCE public.reports_idreport_seq;
DROP TABLE public.reports;
DROP SEQUENCE public.productsizes_idproductsize_seq;
DROP SEQUENCE public.products_idproduct_seq;
DROP VIEW public.productratings;
DROP TABLE public.reviews;
DROP SEQUENCE public.orders_idorder_seq;
DROP SEQUENCE public.orderproducts_idorderproduct_seq;
DROP SEQUENCE public.logs_idlog_seq;
DROP TABLE public.logs;
DROP SEQUENCE public.favorites_idfavorites_seq;
DROP TABLE public.favorites;
DROP VIEW public.dailysales;
DROP TABLE public.productsizes;
DROP TABLE public.products;
DROP TABLE public.orders;
DROP TABLE public.orderproducts;
DROP SEQUENCE public.categories_idcategory_seq;
DROP TABLE public.categories;
DROP SEQUENCE public.brands_idbrand_seq;
DROP TABLE public.brands;
DROP SEQUENCE public.basket_idbasket_seq;
DROP TABLE public.basket;
DROP FUNCTION public.getusertotalspent(user_id integer);
DROP FUNCTION public.gettoprevenueproducts(start_date date, end_date date, limit_count integer);
DROP FUNCTION public.getrevenuebycategory(start_date date, end_date date);
DROP PROCEDURE public.generatetopcustomersreport(IN start_date date, IN end_date date, IN report_name character varying, IN user_id integer);
DROP PROCEDURE public.generatesalesreport(IN start_date date, IN end_date date, IN report_name character varying, IN report_type character varying, IN user_id integer);
DROP PROCEDURE public.generatecategoryrevenuereport(IN start_date date, IN end_date date, IN report_name character varying, IN user_id integer);
DROP FUNCTION public.decreasestock();
DROP FUNCTION public.check_product_quantity();
DROP FUNCTION public.adddefaultsizes();
--
-- TOC entry 264 (class 1255 OID 25267)
-- Name: adddefaultsizes(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.adddefaultsizes() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    s INT;
BEGIN
    FOR s IN 36..45 LOOP
        INSERT INTO ProductSizes(IdProduct, Size, Quantity)
        VALUES (NEW.IdProduct, s, 0);
    END LOOP;
    RETURN NEW;
END;
$$;


--
-- TOC entry 245 (class 1255 OID 33249)
-- Name: check_product_quantity(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.check_product_quantity() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.Quantity < 0 THEN
        RAISE EXCEPTION 'Количество товара не может быть отрицательным (IdProductSize=%)', NEW.IdProductSize;
    END IF;
    RETURN NEW;
END;
$$;


--
-- TOC entry 247 (class 1255 OID 25251)
-- Name: decreasestock(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.decreasestock() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE ProductSizes
    SET Quantity = Quantity - NEW.Quantity
    WHERE IdProductSize = NEW.IdProductSize;
    RETURN NEW;
END;
$$;


--
-- TOC entry 263 (class 1255 OID 25265)
-- Name: generatecategoryrevenuereport(date, date, character varying, integer); Type: PROCEDURE; Schema: public; Owner: -
--

CREATE PROCEDURE public.generatecategoryrevenuereport(IN start_date date, IN end_date date, IN report_name character varying, IN user_id integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO Reports(ReportName, ReportType, ReportData, IdUser)
    SELECT report_name, 
           'CategoryRevenue',
           string_agg(c.CategoryName || ': ' || SUM(op.Quantity * p.Price), ', '),
           user_id
    FROM Orders o
    JOIN OrderProducts op ON o.IdOrder = op.IdOrder
    JOIN ProductSizes ps ON op.IdProductSize = ps.IdProductSize
    JOIN Products p ON ps.IdProduct = p.IdProduct
    JOIN Categories c ON p.IdCategory = c.IdCategory
    WHERE o.OrderDate BETWEEN start_date AND end_date
    GROUP BY user_id;
END;
$$;


--
-- TOC entry 261 (class 1255 OID 25262)
-- Name: generatesalesreport(date, date, character varying, character varying, integer); Type: PROCEDURE; Schema: public; Owner: -
--

CREATE PROCEDURE public.generatesalesreport(IN start_date date, IN end_date date, IN report_name character varying, IN report_type character varying, IN user_id integer)
    LANGUAGE plpgsql
    AS $$
DECLARE
    total_sales DECIMAL;
    total_orders INT;
BEGIN
    IF report_type = 'Sales' THEN
        SELECT COALESCE(SUM(op.Quantity * p.Price),0)
        INTO total_sales
        FROM Orders o
        JOIN OrderProducts op ON o.IdOrder = op.IdOrder
        JOIN ProductSizes ps ON op.IdProductSize = ps.IdProductSize
        JOIN Products p ON ps.IdProduct = p.IdProduct
        WHERE o.OrderDate BETWEEN start_date AND end_date;

        INSERT INTO Reports(ReportName, ReportType, ReportData, IdUser)
        VALUES (report_name, report_type, 'Total Sales: ' || total_sales, user_id);

    ELSIF report_type = 'Orders' THEN
        SELECT COUNT(*)
        INTO total_orders
        FROM Orders
        WHERE OrderDate BETWEEN start_date AND end_date;

        INSERT INTO Reports(ReportName, ReportType, ReportData, IdUser)
        VALUES (report_name, report_type, 'Total Orders: ' || total_orders, user_id);

    END IF;
END;
$$;


--
-- TOC entry 262 (class 1255 OID 25266)
-- Name: generatetopcustomersreport(date, date, character varying, integer); Type: PROCEDURE; Schema: public; Owner: -
--

CREATE PROCEDURE public.generatetopcustomersreport(IN start_date date, IN end_date date, IN report_name character varying, IN user_id integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO Reports(ReportName, ReportType, ReportData, IdUser)
    SELECT report_name,
           'TopCustomers',
           string_agg(u.FullName || ': ' || SUM(op.Quantity * p.Price), ', '),
           user_id
    FROM Orders o
    JOIN OrderProducts op ON o.IdOrder = op.IdOrder
    JOIN ProductSizes ps ON op.IdProductSize = ps.IdProductSize
    JOIN Products p ON ps.IdProduct = p.IdProduct
    JOIN Users u ON o.IdUser = u.IdUser
    WHERE o.OrderDate BETWEEN start_date AND end_date
    GROUP BY user_id
    ORDER BY SUM(op.Quantity * p.Price) DESC
    LIMIT 10;
END;
$$;


--
-- TOC entry 260 (class 1255 OID 25264)
-- Name: getrevenuebycategory(date, date); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.getrevenuebycategory(start_date date, end_date date) RETURNS TABLE(categoryname character varying, totalrevenue numeric, totalquantity integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT c.CategoryName,
           SUM(op.Quantity * p.Price) AS TotalRevenue,
           SUM(op.Quantity) AS TotalQuantity
    FROM Orders o
    JOIN OrderProducts op ON o.IdOrder = op.IdOrder
    JOIN ProductSizes ps ON op.IdProductSize = ps.IdProductSize
    JOIN Products p ON ps.IdProduct = p.IdProduct
    JOIN Categories c ON p.IdCategory = c.IdCategory
    WHERE o.OrderDate BETWEEN start_date AND end_date
    GROUP BY c.CategoryName
    ORDER BY TotalRevenue DESC;
END;
$$;


--
-- TOC entry 259 (class 1255 OID 25263)
-- Name: gettoprevenueproducts(date, date, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.gettoprevenueproducts(start_date date, end_date date, limit_count integer) RETURNS TABLE(productname character varying, totalrevenue numeric, totalquantity integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT p.Name,
           SUM(op.Quantity * p.Price) AS TotalRevenue,
           SUM(op.Quantity) AS TotalQuantity
    FROM Orders o
    JOIN OrderProducts op ON o.IdOrder = op.IdOrder
    JOIN ProductSizes ps ON op.IdProductSize = ps.IdProductSize
    JOIN Products p ON ps.IdProduct = p.IdProduct
    WHERE o.OrderDate BETWEEN start_date AND end_date
    GROUP BY p.Name
    ORDER BY TotalRevenue DESC
    LIMIT limit_count;
END;
$$;


--
-- TOC entry 246 (class 1255 OID 25250)
-- Name: getusertotalspent(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.getusertotalspent(user_id integer) RETURNS numeric
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN (
        SELECT COALESCE(SUM(op.Quantity * p.Price), 0)
        FROM Orders o
        JOIN OrderProducts op ON o.IdOrder = op.IdOrder
        JOIN ProductSizes ps ON op.IdProductSize = ps.IdProductSize
        JOIN Products p ON ps.IdProduct = p.IdProduct
        WHERE o.IdUser = user_id
    );
END;
$$;


SET default_table_access_method = heap;

--
-- TOC entry 228 (class 1259 OID 25115)
-- Name: basket; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.basket (
    idbasket integer NOT NULL,
    iduser integer NOT NULL,
    idproductsize integer NOT NULL,
    quantity integer NOT NULL
);


--
-- TOC entry 227 (class 1259 OID 25114)
-- Name: basket_idbasket_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.basket_idbasket_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5060 (class 0 OID 0)
-- Dependencies: 227
-- Name: basket_idbasket_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.basket_idbasket_seq OWNED BY public.basket.idbasket;


--
-- TOC entry 220 (class 1259 OID 25069)
-- Name: brands; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.brands (
    idbrand integer NOT NULL,
    brandname character varying(100) NOT NULL
);


--
-- TOC entry 219 (class 1259 OID 25068)
-- Name: brands_idbrand_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.brands_idbrand_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5061 (class 0 OID 0)
-- Dependencies: 219
-- Name: brands_idbrand_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.brands_idbrand_seq OWNED BY public.brands.idbrand;


--
-- TOC entry 222 (class 1259 OID 25077)
-- Name: categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.categories (
    idcategory integer NOT NULL,
    categoryname character varying(100) NOT NULL
);


--
-- TOC entry 221 (class 1259 OID 25076)
-- Name: categories_idcategory_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.categories_idcategory_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5062 (class 0 OID 0)
-- Dependencies: 221
-- Name: categories_idcategory_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.categories_idcategory_seq OWNED BY public.categories.idcategory;


--
-- TOC entry 236 (class 1259 OID 25183)
-- Name: orderproducts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orderproducts (
    idorderproduct integer NOT NULL,
    idorder integer,
    idproductsize integer,
    quantity integer NOT NULL
);


--
-- TOC entry 234 (class 1259 OID 25170)
-- Name: orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orders (
    idorder integer NOT NULL,
    iduser integer,
    orderdate timestamp without time zone DEFAULT now()
);


--
-- TOC entry 224 (class 1259 OID 25084)
-- Name: products; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.products (
    idproduct integer NOT NULL,
    name character varying(100) NOT NULL,
    imageurl character varying(255),
    price numeric NOT NULL,
    idbrand integer NOT NULL,
    idcategory integer NOT NULL
);


--
-- TOC entry 226 (class 1259 OID 25103)
-- Name: productsizes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.productsizes (
    idproductsize integer NOT NULL,
    idproduct integer NOT NULL,
    size integer NOT NULL,
    quantity integer NOT NULL
);


--
-- TOC entry 242 (class 1259 OID 25235)
-- Name: dailysales; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dailysales AS
 SELECT date(o.orderdate) AS day,
    sum(((op.quantity)::numeric * p.price)) AS revenue
   FROM (((public.orders o
     JOIN public.orderproducts op ON ((o.idorder = op.idorder)))
     JOIN public.productsizes ps ON ((op.idproductsize = ps.idproductsize)))
     JOIN public.products p ON ((ps.idproduct = p.idproduct)))
  GROUP BY (date(o.orderdate))
  ORDER BY (date(o.orderdate)) DESC;


--
-- TOC entry 230 (class 1259 OID 25132)
-- Name: favorites; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.favorites (
    idfavorites integer NOT NULL,
    iduser integer NOT NULL,
    idproductsize integer NOT NULL
);


--
-- TOC entry 229 (class 1259 OID 25131)
-- Name: favorites_idfavorites_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.favorites_idfavorites_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5063 (class 0 OID 0)
-- Dependencies: 229
-- Name: favorites_idfavorites_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.favorites_idfavorites_seq OWNED BY public.favorites.idfavorites;


--
-- TOC entry 240 (class 1259 OID 25215)
-- Name: logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.logs (
    idlog integer NOT NULL,
    iduser integer NOT NULL,
    action character varying(100) NOT NULL,
    entity character varying(50),
    entityid integer,
    details text,
    createdat timestamp without time zone DEFAULT now() NOT NULL
);


--
-- TOC entry 239 (class 1259 OID 25214)
-- Name: logs_idlog_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.logs_idlog_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5064 (class 0 OID 0)
-- Dependencies: 239
-- Name: logs_idlog_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.logs_idlog_seq OWNED BY public.logs.idlog;


--
-- TOC entry 235 (class 1259 OID 25182)
-- Name: orderproducts_idorderproduct_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.orderproducts_idorderproduct_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5065 (class 0 OID 0)
-- Dependencies: 235
-- Name: orderproducts_idorderproduct_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.orderproducts_idorderproduct_seq OWNED BY public.orderproducts.idorderproduct;


--
-- TOC entry 233 (class 1259 OID 25169)
-- Name: orders_idorder_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.orders_idorder_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5066 (class 0 OID 0)
-- Dependencies: 233
-- Name: orders_idorder_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.orders_idorder_seq OWNED BY public.orders.idorder;


--
-- TOC entry 232 (class 1259 OID 25149)
-- Name: reviews; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reviews (
    idreview integer NOT NULL,
    idproduct integer NOT NULL,
    rating integer NOT NULL,
    comment text,
    reviewdate timestamp without time zone DEFAULT now() NOT NULL,
    iduser integer,
    CONSTRAINT reviews_rating_check CHECK (((rating >= 1) AND (rating <= 5)))
);


--
-- TOC entry 243 (class 1259 OID 25240)
-- Name: productratings; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.productratings AS
 SELECT p.name,
    round(avg(r.rating), 2) AS avgrating,
    count(r.idreview) AS reviewcount
   FROM (public.products p
     LEFT JOIN public.reviews r ON ((p.idproduct = r.idproduct)))
  GROUP BY p.name;


--
-- TOC entry 223 (class 1259 OID 25083)
-- Name: products_idproduct_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.products_idproduct_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5067 (class 0 OID 0)
-- Dependencies: 223
-- Name: products_idproduct_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.products_idproduct_seq OWNED BY public.products.idproduct;


--
-- TOC entry 225 (class 1259 OID 25102)
-- Name: productsizes_idproductsize_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.productsizes_idproductsize_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5068 (class 0 OID 0)
-- Dependencies: 225
-- Name: productsizes_idproductsize_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.productsizes_idproductsize_seq OWNED BY public.productsizes.idproductsize;


--
-- TOC entry 238 (class 1259 OID 25200)
-- Name: reports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reports (
    idreport integer NOT NULL,
    reportname character varying(255) NOT NULL,
    reporttype character varying(50) NOT NULL,
    reportdata text NOT NULL,
    iduser integer NOT NULL,
    createdat timestamp without time zone DEFAULT now() NOT NULL
);


--
-- TOC entry 237 (class 1259 OID 25199)
-- Name: reports_idreport_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reports_idreport_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5069 (class 0 OID 0)
-- Dependencies: 237
-- Name: reports_idreport_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reports_idreport_seq OWNED BY public.reports.idreport;


--
-- TOC entry 231 (class 1259 OID 25148)
-- Name: reviews_idreview_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reviews_idreview_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5070 (class 0 OID 0)
-- Dependencies: 231
-- Name: reviews_idreview_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reviews_idreview_seq OWNED BY public.reviews.idreview;


--
-- TOC entry 218 (class 1259 OID 25058)
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    iduser integer NOT NULL,
    fullname character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    passwordhash character varying(255) NOT NULL,
    roleid integer NOT NULL
);


--
-- TOC entry 244 (class 1259 OID 25244)
-- Name: topcustomers; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.topcustomers AS
 SELECT u.fullname,
    u.email,
    sum(((op.quantity)::numeric * p.price)) AS totalspent
   FROM ((((public.orders o
     JOIN public.orderproducts op ON ((o.idorder = op.idorder)))
     JOIN public.productsizes ps ON ((op.idproductsize = ps.idproductsize)))
     JOIN public.products p ON ((ps.idproduct = p.idproduct)))
     JOIN public.users u ON ((o.iduser = u.iduser)))
  GROUP BY u.iduser, u.fullname, u.email
  ORDER BY (sum(((op.quantity)::numeric * p.price))) DESC
 LIMIT 10;


--
-- TOC entry 241 (class 1259 OID 25230)
-- Name: topproducts; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.topproducts AS
 SELECT p.name,
    sum(op.quantity) AS totalsold
   FROM ((public.orderproducts op
     JOIN public.productsizes ps ON ((op.idproductsize = ps.idproductsize)))
     JOIN public.products p ON ((ps.idproduct = p.idproduct)))
  GROUP BY p.name
  ORDER BY (sum(op.quantity)) DESC
 LIMIT 10;


--
-- TOC entry 217 (class 1259 OID 25057)
-- Name: users_iduser_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_iduser_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5071 (class 0 OID 0)
-- Dependencies: 217
-- Name: users_iduser_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_iduser_seq OWNED BY public.users.iduser;


--
-- TOC entry 4827 (class 2604 OID 25118)
-- Name: basket idbasket; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.basket ALTER COLUMN idbasket SET DEFAULT nextval('public.basket_idbasket_seq'::regclass);


--
-- TOC entry 4823 (class 2604 OID 25072)
-- Name: brands idbrand; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.brands ALTER COLUMN idbrand SET DEFAULT nextval('public.brands_idbrand_seq'::regclass);


--
-- TOC entry 4824 (class 2604 OID 25080)
-- Name: categories idcategory; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories ALTER COLUMN idcategory SET DEFAULT nextval('public.categories_idcategory_seq'::regclass);


--
-- TOC entry 4828 (class 2604 OID 25135)
-- Name: favorites idfavorites; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorites ALTER COLUMN idfavorites SET DEFAULT nextval('public.favorites_idfavorites_seq'::regclass);


--
-- TOC entry 4836 (class 2604 OID 25218)
-- Name: logs idlog; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs ALTER COLUMN idlog SET DEFAULT nextval('public.logs_idlog_seq'::regclass);


--
-- TOC entry 4833 (class 2604 OID 25186)
-- Name: orderproducts idorderproduct; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orderproducts ALTER COLUMN idorderproduct SET DEFAULT nextval('public.orderproducts_idorderproduct_seq'::regclass);


--
-- TOC entry 4831 (class 2604 OID 25173)
-- Name: orders idorder; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders ALTER COLUMN idorder SET DEFAULT nextval('public.orders_idorder_seq'::regclass);


--
-- TOC entry 4825 (class 2604 OID 25087)
-- Name: products idproduct; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products ALTER COLUMN idproduct SET DEFAULT nextval('public.products_idproduct_seq'::regclass);


--
-- TOC entry 4826 (class 2604 OID 25106)
-- Name: productsizes idproductsize; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.productsizes ALTER COLUMN idproductsize SET DEFAULT nextval('public.productsizes_idproductsize_seq'::regclass);


--
-- TOC entry 4834 (class 2604 OID 25203)
-- Name: reports idreport; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports ALTER COLUMN idreport SET DEFAULT nextval('public.reports_idreport_seq'::regclass);


--
-- TOC entry 4829 (class 2604 OID 25152)
-- Name: reviews idreview; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews ALTER COLUMN idreview SET DEFAULT nextval('public.reviews_idreview_seq'::regclass);


--
-- TOC entry 4822 (class 2604 OID 25061)
-- Name: users iduser; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN iduser SET DEFAULT nextval('public.users_iduser_seq'::regclass);


--
-- TOC entry 5042 (class 0 OID 25115)
-- Dependencies: 228
-- Data for Name: basket; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.basket (idbasket, iduser, idproductsize, quantity) FROM stdin;
1	3	1	1
2	3	2	2
3	4	5	1
5	16	61	2
\.


--
-- TOC entry 5034 (class 0 OID 25069)
-- Dependencies: 220
-- Data for Name: brands; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.brands (idbrand, brandname) FROM stdin;
1	Nike
2	Adidas
3	Puma
4	Reebok
\.


--
-- TOC entry 5036 (class 0 OID 25077)
-- Dependencies: 222
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.categories (idcategory, categoryname) FROM stdin;
1	Кроссовки
2	Ботинки
3	Кеды
5	mes
\.


--
-- TOC entry 5044 (class 0 OID 25132)
-- Dependencies: 230
-- Data for Name: favorites; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.favorites (idfavorites, iduser, idproductsize) FROM stdin;
1	3	1
2	4	2
3	5	3
6	16	70
9	24	50
\.


--
-- TOC entry 5054 (class 0 OID 25215)
-- Dependencies: 240
-- Data for Name: logs; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.logs (idlog, iduser, action, entity, entityid, details, createdat) FROM stdin;
14	18	CREATE	user/reports	\N	user CREATE reports	2025-10-17 12:29:49.957029
15	17	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 00:45:55.423885
16	17	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 01:00:17.53351
17	16	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 01:01:31.402638
18	17	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 01:02:05.711891
19	17	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 01:10:55.391486
20	17	CREATE	backup	0	Создан бэкап базы данных: shoes_store_backup_20251020_011348.sql (0.04 MB)	2025-10-20 01:13:48.812348
21	17	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 01:13:48.184352
22	17	CREATE	backup	0	Создан бэкап базы данных: shoes_store_backup_20251020_013301.sql (0.04 MB)	2025-10-20 01:33:02.052406
23	17	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 01:33:01.440135
24	16	ADD_TO_BASKET	basket	4	Добавлен товар в корзину пользователя ID: 16	2025-10-26 16:48:41.441655
25	16	CREATE	user/basket	\N	user CREATE basket	2025-10-26 16:48:41.4089
26	16	ADD_TO_FAVORITES	favorite	4	Добавлен товар в избранное пользователя ID: 16	2025-10-26 16:48:43.607863
27	16	CREATE	user/favorites	\N	user CREATE favorites	2025-10-26 16:48:43.573866
28	16	UPDATE_BASKET	basket	4	Обновлено количество в корзине ID: 4	2025-10-26 16:48:53.035521
29	16	UPDATE	user/4	\N	user UPDATE 4	2025-10-26 16:48:53.001182
30	16	UPDATE_BASKET	basket	4	Обновлено количество в корзине ID: 4	2025-10-26 16:48:53.437099
31	16	UPDATE	user/4	\N	user UPDATE 4	2025-10-26 16:48:53.406735
32	16	UPDATE_BASKET	basket	4	Обновлено количество в корзине ID: 4	2025-10-26 16:48:53.638659
33	16	UPDATE	user/4	\N	user UPDATE 4	2025-10-26 16:48:53.608349
34	16	UPDATE_BASKET	basket	4	Обновлено количество в корзине ID: 4	2025-10-26 16:48:53.832194
35	16	UPDATE	user/4	\N	user UPDATE 4	2025-10-26 16:48:53.801709
36	16	UPDATE_BASKET	basket	4	Обновлено количество в корзине ID: 4	2025-10-26 16:48:54.001305
37	16	UPDATE	user/4	\N	user UPDATE 4	2025-10-26 16:48:53.97101
38	16	UPDATE_BASKET	basket	4	Обновлено количество в корзине ID: 4	2025-10-26 16:48:54.172539
39	16	UPDATE	user/4	\N	user UPDATE 4	2025-10-26 16:48:54.141962
40	16	UPDATE_BASKET	basket	4	Обновлено количество в корзине ID: 4	2025-10-26 16:48:54.339695
41	16	UPDATE	user/4	\N	user UPDATE 4	2025-10-26 16:48:54.30913
42	16	UPDATE_BASKET	basket	4	Обновлено количество в корзине ID: 4	2025-10-26 16:48:54.487223
43	16	UPDATE	user/4	\N	user UPDATE 4	2025-10-26 16:48:54.479267
44	16	UPDATE_BASKET	basket	4	Обновлено количество в корзине ID: 4	2025-10-26 16:48:54.641464
45	16	UPDATE	user/4	\N	user UPDATE 4	2025-10-26 16:48:54.636695
46	16	UPDATE_BASKET	basket	4	Обновлено количество в корзине ID: 4	2025-10-26 16:48:54.844631
47	16	UPDATE	user/4	\N	user UPDATE 4	2025-10-26 16:48:54.814244
48	16	UPDATE_BASKET	basket	4	Обновлено количество в корзине ID: 4	2025-10-26 16:48:55.022592
49	16	UPDATE	user/4	\N	user UPDATE 4	2025-10-26 16:48:54.991891
50	16	UPDATE_BASKET	basket	4	Обновлено количество в корзине ID: 4	2025-10-26 16:48:55.238359
51	16	UPDATE	user/4	\N	user UPDATE 4	2025-10-26 16:48:55.207446
52	16	REMOVE_FROM_BASKET	basket	4	Удален товар из корзины ID: 4	2025-10-26 16:48:56.153574
53	16	DELETE	user/4	\N	user DELETE 4	2025-10-26 16:48:56.122483
54	16	ADD_TO_BASKET	basket	5	Добавлен товар в корзину пользователя ID: 16	2025-10-26 17:00:27.010349
55	16	CREATE	user/basket	\N	user CREATE basket	2025-10-26 17:00:26.985857
56	16	UPDATE_BASKET	basket	5	Обновлено количество в корзине ID: 5	2025-10-26 18:50:31.829088
57	16	UPDATE	user/5	\N	user UPDATE 5	2025-10-26 18:50:31.814661
58	16	UPDATE_BASKET	basket	5	Обновлено количество в корзине ID: 5	2025-10-26 18:50:33.366077
59	16	UPDATE	user/5	\N	user UPDATE 5	2025-10-26 18:50:33.350391
60	16	ADD_TO_FAVORITES	favorite	5	Добавлен товар в избранное пользователя ID: 16	2025-10-26 19:03:54.307508
61	16	CREATE	user/favorites	\N	user CREATE favorites	2025-10-26 19:03:54.301305
62	16	REMOVE_FROM_FAVORITES	favorite	5	Удален товар из избранного ID: 5	2025-10-26 19:04:01.350477
63	16	DELETE	user/5	\N	user DELETE 5	2025-10-26 19:04:01.345962
64	16	REMOVE_FROM_FAVORITES	favorite	4	Удален товар из избранного ID: 4	2025-10-26 19:04:02.488802
65	16	DELETE	user/4	\N	user DELETE 4	2025-10-26 19:04:02.484317
66	16	UPDATE_BASKET	basket	5	Обновлено количество в корзине ID: 5	2025-10-26 21:17:09.138929
67	16	UPDATE	user/5	\N	user UPDATE 5	2025-10-26 21:17:09.133258
68	16	ADD_TO_FAVORITES	favorite	6	Добавлен товар в избранное пользователя ID: 16	2025-10-26 21:41:40.743306
69	16	CREATE	user/favorites	\N	user CREATE favorites	2025-10-26 21:41:40.737858
70	17	DELETE	user	13	Удален пользователь с ID: 13	2025-10-26 21:55:06.031809
71	17	DELETE	admin/users	\N	admin DELETE users	2025-10-26 21:55:06.009688
72	17	DELETE	user	19	Удален пользователь с ID: 19	2025-10-26 21:55:20.117669
73	17	DELETE	admin/users	\N	admin DELETE users	2025-10-26 21:55:20.113129
74	17	DELETE	user	20	Удален пользователь с ID: 20	2025-10-26 21:58:14.17321
75	17	DELETE	admin/users	\N	admin DELETE users	2025-10-26 21:58:14.15314
76	17	DELETE	user	21	Удален пользователь с ID: 21	2025-10-26 22:03:36.007548
77	17	DELETE	admin/users	\N	admin DELETE users	2025-10-26 22:03:35.993212
78	17	DELETE	user	23	Удален пользователь с ID: 23	2025-10-26 22:05:15.012806
79	17	DELETE	admin/users	\N	admin DELETE users	2025-10-26 22:05:15.007784
80	24	ADD_TO_BASKET	basket	6	Добавлен товар в корзину пользователя ID: 24	2025-10-26 22:06:10.490284
81	24	CREATE	user/basket	\N	user CREATE basket	2025-10-26 22:06:10.480958
82	24	ADD_TO_FAVORITES	favorite	7	Добавлен товар в избранное пользователя ID: 24	2025-10-26 22:06:14.168181
83	24	CREATE	user/favorites	\N	user CREATE favorites	2025-10-26 22:06:14.164972
84	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:11.478533
85	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:11.47524
86	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:12.937403
87	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:12.935746
88	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:14.388239
89	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:14.383732
90	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:16.245321
91	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:16.24088
92	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:19.559666
93	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:19.554951
94	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:29.627523
95	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:29.623109
96	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:31.800545
97	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:31.79621
98	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:37.32866
99	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:37.325979
100	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:38.238663
101	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:38.23707
102	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:38.429551
103	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:38.428277
104	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:38.651812
105	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:38.648851
106	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:38.791489
107	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:38.79008
108	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:38.963116
109	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:38.961778
110	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:39.135718
111	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:39.13476
112	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:39.481959
113	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:39.477709
114	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:39.693938
115	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:39.689723
116	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:39.936708
117	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:39.932327
118	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:40.136826
119	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:40.135451
120	24	UPDATE_BASKET	basket	6	Обновлено количество в корзине ID: 6	2025-10-26 22:08:40.865746
121	24	UPDATE	user/6	\N	user UPDATE 6	2025-10-26 22:08:40.861725
122	24	REMOVE_FROM_BASKET	basket	6	Удален товар из корзины ID: 6	2025-10-26 22:08:41.68882
123	24	DELETE	user/6	\N	user DELETE 6	2025-10-26 22:08:41.684415
124	24	ADD_TO_BASKET	basket	7	Добавлен товар в корзину пользователя ID: 24	2025-10-26 22:08:47.615439
125	24	CREATE	user/basket	\N	user CREATE basket	2025-10-26 22:08:47.610835
126	24	ADD_TO_FAVORITES	favorite	8	Добавлен товар в избранное пользователя ID: 24	2025-10-26 22:08:50.958658
127	24	CREATE	user/favorites	\N	user CREATE favorites	2025-10-26 22:08:50.956388
128	24	ADD_TO_BASKET	basket	8	Добавлен товар в корзину пользователя ID: 24	2025-10-27 14:05:30.244024
129	24	CREATE	user/basket	\N	user CREATE basket	2025-10-27 14:05:30.221213
130	24	ADD_TO_BASKET	basket	9	Добавлен товар в корзину пользователя ID: 24	2025-10-27 14:05:37.413705
131	24	CREATE	user/basket	\N	user CREATE basket	2025-10-27 14:05:37.383091
132	24	UPDATE_BASKET	basket	8	Обновлено количество в корзине ID: 8	2025-10-27 14:05:42.374006
133	24	UPDATE	user/8	\N	user UPDATE 8	2025-10-27 14:05:42.367598
134	24	UPDATE_BASKET	basket	9	Обновлено количество в корзине ID: 9	2025-10-27 14:05:42.574291
135	24	UPDATE	user/9	\N	user UPDATE 9	2025-10-27 14:05:42.544373
136	24	UPDATE_BASKET	basket	8	Обновлено количество в корзине ID: 8	2025-10-27 14:05:42.753229
137	24	UPDATE	user/8	\N	user UPDATE 8	2025-10-27 14:05:42.722427
138	24	UPDATE_BASKET	basket	9	Обновлено количество в корзине ID: 9	2025-10-27 14:05:43.163064
139	24	UPDATE	user/9	\N	user UPDATE 9	2025-10-27 14:05:43.132527
140	24	REMOVE_FROM_BASKET	basket	8	Удален товар из корзины ID: 8	2025-10-27 14:05:43.76004
141	24	DELETE	user/8	\N	user DELETE 8	2025-10-27 14:05:43.753425
142	24	REMOVE_FROM_BASKET	basket	9	Удален товар из корзины ID: 9	2025-10-27 14:05:44.214725
143	24	DELETE	user/9	\N	user DELETE 9	2025-10-27 14:05:44.184515
144	24	REMOVE_FROM_BASKET	basket	7	Удален товар из корзины ID: 7	2025-10-27 14:05:45.156202
145	24	DELETE	user/7	\N	user DELETE 7	2025-10-27 14:05:45.126035
146	17	CREATE	category	5	Создана категория: mes	2025-10-27 14:14:58.508668
147	17	CREATE	admin/categories	\N	admin CREATE categories	2025-10-27 14:14:58.471503
148	24	CREATE	user/change	\N	user CREATE change	2025-10-28 14:06:32.507732
149	24	CREATE	user/change	\N	user CREATE change	2025-10-28 14:09:33.376267
150	24	CREATE	user/change	\N	user CREATE change	2025-10-28 14:09:51.254461
151	24	CREATE	user/change	\N	user CREATE change	2025-10-28 14:09:54.94287
152	24	CREATE	user/change	\N	user CREATE change	2025-10-28 14:10:33.782537
153	24	CREATE	user/change	\N	user CREATE change	2025-10-28 14:15:55.384424
154	24	PASSWORD_CHANGE	user	24	Изменен пароль для пользователя: rmusaevdg5@gmail.com	2025-10-28 14:16:19.288129
155	24	CREATE	user/change	\N	user CREATE change	2025-10-28 14:16:19.153814
156	24	ADD_TO_BASKET	basket	10	Добавлен товар в корзину пользователя ID: 24	2025-10-28 14:32:14.267444
157	24	CREATE	user/basket	\N	user CREATE basket	2025-10-28 14:32:14.257809
158	24	UPDATE_BASKET	basket	10	Обновлено количество в корзине ID: 10	2025-10-28 14:32:17.294817
159	24	UPDATE	user/10	\N	user UPDATE 10	2025-10-28 14:32:17.263987
160	24	CREATE	order	3	Создан заказ для пользователя ID: 24	2025-10-28 14:32:21.724943
161	24	CREATE	user/orders	\N	user CREATE orders	2025-10-28 14:32:20.181991
162	24	CREATE	user/order-products	\N	user CREATE order-products	2025-10-28 14:32:21.766771
163	24	REMOVE_FROM_BASKET	basket	10	Удален товар из корзины ID: 10	2025-10-28 14:32:21.817763
164	24	DELETE	user/10	\N	user DELETE 10	2025-10-28 14:32:21.813662
165	24	CREATE	review	7	Создан отзыв для товара ID: 7 пользователем ID: 24	2025-10-28 15:04:21.677751
166	24	CREATE	user/reviews	\N	user CREATE reviews	2025-10-28 15:04:21.583508
167	24	CREATE	user/reviews	\N	user CREATE reviews	2025-10-28 15:04:36.154484
168	24	UPDATE	review	7	Обновлен отзыв ID: 7	2025-10-28 15:05:56.078431
169	24	UPDATE	user/7	\N	user UPDATE 7	2025-10-28 15:05:56.04425
170	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-10-29 14:04:46.525119
171	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-10-29 14:07:40.742661
172	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-10-29 14:07:59.855638
173	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по инвентарю	2025-10-29 14:08:11.560311
174	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-10-29 14:08:44.357057
175	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-10-29 14:08:55.752827
176	18	GENERATE_TEXT	report	0	Сгенерирован текстовый отчет по клиентам	2025-10-29 14:09:20.537731
177	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-10-29 14:09:43.91053
178	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-10-29 14:17:39.004504
179	24	ADD_TO_BASKET	basket	11	Добавлен товар в корзину пользователя ID: 24	2025-10-29 14:18:18.288498
180	24	CREATE	user/basket	\N	user CREATE basket	2025-10-29 14:18:18.285148
181	17	UPDATE	product	5	Обновлен товар: Nike Air Force 1	2025-10-30 12:03:19.355287
182	17	UPDATE	admin/products	\N	admin UPDATE products	2025-10-30 12:03:19.336613
183	24	CREATE	user/reviews	\N	user CREATE reviews	2025-10-30 12:13:45.667958
184	24	REMOVE_FROM_FAVORITES	favorite	8	Удален товар из избранного ID: 8	2025-10-30 12:44:58.309723
185	24	DELETE	user/8	\N	user DELETE 8	2025-10-30 12:44:58.277079
186	24	REMOVE_FROM_FAVORITES	favorite	7	Удален товар из избранного ID: 7	2025-10-30 12:44:58.829493
187	24	DELETE	user/7	\N	user DELETE 7	2025-10-30 12:44:58.798537
188	24	CREATE	user/reviews	\N	user CREATE reviews	2025-10-30 12:45:53.355042
189	24	ADD_TO_FAVORITES	favorite	9	Добавлен товар в избранное пользователя ID: 24	2025-10-30 12:47:28.222955
190	24	CREATE	user/favorites	\N	user CREATE favorites	2025-10-30 12:47:28.190606
191	24	ADD_TO_BASKET	basket	12	Добавлен товар в корзину пользователя ID: 24	2025-10-30 12:47:33.431111
192	24	CREATE	user/basket	\N	user CREATE basket	2025-10-30 12:47:33.395357
193	24	UPDATE_BASKET	basket	12	Обновлено количество в корзине ID: 12	2025-10-30 12:47:37.659125
194	24	UPDATE	user/12	\N	user UPDATE 12	2025-10-30 12:47:37.631427
195	24	UPDATE_BASKET	basket	12	Обновлено количество в корзине ID: 12	2025-10-30 12:47:37.888411
196	24	UPDATE	user/12	\N	user UPDATE 12	2025-10-30 12:47:37.860348
197	24	UPDATE_BASKET	basket	12	Обновлено количество в корзине ID: 12	2025-10-30 12:47:38.470044
198	24	UPDATE	user/12	\N	user UPDATE 12	2025-10-30 12:47:38.439398
199	24	UPDATE_BASKET	basket	12	Обновлено количество в корзине ID: 12	2025-10-30 12:47:38.817063
200	24	UPDATE	user/12	\N	user UPDATE 12	2025-10-30 12:47:38.78645
201	24	REMOVE_FROM_BASKET	basket	12	Удален товар из корзины ID: 12	2025-10-30 12:47:39.231933
202	24	DELETE	user/12	\N	user DELETE 12	2025-10-30 12:47:39.200652
203	24	REMOVE_FROM_BASKET	basket	11	Удален товар из корзины ID: 11	2025-10-30 12:47:55.818213
204	24	DELETE	user/11	\N	user DELETE 11	2025-10-30 12:47:55.788079
205	24	ADD_TO_BASKET	basket	13	Добавлен товар в корзину пользователя ID: 24	2025-10-30 12:47:59.181633
206	24	CREATE	user/basket	\N	user CREATE basket	2025-10-30 12:47:59.151054
207	24	CREATE	order	4	Создан заказ для пользователя ID: 24	2025-10-30 12:49:43.075047
208	24	CREATE	user/orders	\N	user CREATE orders	2025-10-30 12:49:39.169725
209	24	CREATE	user/order-products	\N	user CREATE order-products	2025-10-30 12:49:43.110265
210	24	REMOVE_FROM_BASKET	basket	13	Удален товар из корзины ID: 13	2025-10-30 12:49:43.153024
211	24	DELETE	user/13	\N	user DELETE 13	2025-10-30 12:49:43.15174
212	24	CREATE	user/change	\N	user CREATE change	2025-10-30 12:50:56.421104
213	24	PASSWORD_CHANGE	user	24	Изменен пароль для пользователя: rmusaevdg5@gmail.com	2025-10-30 12:51:17.509442
214	24	CREATE	user/change	\N	user CREATE change	2025-10-30 12:51:17.385757
215	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-10-30 12:56:58.560922
216	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по инвентарю	2025-10-30 12:57:24.381788
217	24	ADD_TO_BASKET	basket	14	Добавлен товар в корзину пользователя ID: 24	2025-10-30 13:48:53.167084
218	24	CREATE	user/basket	\N	user CREATE basket	2025-10-30 13:48:53.123615
219	24	CREATE	user/reviews	\N	user CREATE reviews	2025-10-30 13:51:59.362845
220	24	UPDATE	review	7	Обновлен отзыв ID: 7	2025-10-30 14:01:41.659967
221	24	UPDATE	user/7	\N	user UPDATE 7	2025-10-30 14:01:41.622066
222	17	UPDATE	product	5	Обновлен товар: Nike Air Force 1	2025-10-30 14:43:55.804135
223	17	UPDATE	admin/products	\N	admin UPDATE products	2025-10-30 14:43:55.768665
224	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-01 14:21:23.196234
225	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по инвентарю	2025-11-01 14:36:05.408907
226	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по инвентарю	2025-11-01 14:36:10.703861
295	24	CREATE	user/orders	\N	user CREATE orders	2025-11-10 16:27:49.593918
227	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по клиентам	2025-11-01 14:52:06.818145
228	24	CREATE	user/reviews	\N	user CREATE reviews	2025-11-01 15:14:34.812109
229	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-11-01 15:15:37.465954
230	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-01 15:15:48.875381
231	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по инвентарю	2025-11-01 15:15:52.764737
232	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по клиентам	2025-11-01 15:15:56.866775
233	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по категориям	2025-11-01 15:16:01.135575
234	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-01 15:16:15.800106
235	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-01 15:16:17.423899
236	17	UPDATE	product	7	Обновлен товар: Adidas Samba	2025-11-03 05:15:19.172783
237	17	UPDATE	admin/products	\N	admin UPDATE products	2025-11-03 05:15:19.150236
238	17	UPDATE	product	7	Обновлен товар: Adidas Samba	2025-11-03 05:16:01.112734
239	17	UPDATE	admin/products	\N	admin UPDATE products	2025-11-03 05:16:01.108414
240	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-11-03 05:16:30.470615
241	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-03 05:16:34.833501
242	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по инвентарю	2025-11-03 05:16:38.32217
243	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по клиентам	2025-11-03 05:16:41.742966
244	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-11-03 05:16:44.950878
245	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-03 05:16:49.568277
246	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-03 05:16:50.976745
247	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-03 05:16:54.277899
248	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-03 05:16:55.752937
249	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-03 05:16:55.935373
250	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по категориям	2025-11-03 05:19:08.421072
251	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по клиентам	2025-11-03 05:19:36.745103
252	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-11-03 05:19:45.339236
253	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-03 05:19:52.428361
254	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-03 05:29:29.327523
255	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-03 05:30:02.015591
256	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-03 05:30:05.716459
257	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-03 05:30:53.459837
258	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-11-03 05:33:27.267687
259	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-06 13:30:55.244586
260	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по продажам	2025-11-06 13:32:47.056693
261	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по категориям	2025-11-08 15:36:48.278426
262	18	GENERATE_EXCEL	report	0	Сгенерирован Excel отчет по инвентарю	2025-11-08 17:58:50.517154
263	17	CREATE	product	8	Создан товар: ewds	2025-11-08 18:59:00.865087
264	17	CREATE	admin/products	\N	admin CREATE products	2025-11-08 18:59:00.810535
265	18	GENERATE_PDF	report	0	Сгенерирован PDF отчет по продажам	2025-11-08 19:38:00.909118
266	17	DELETE	admin/products	\N	admin DELETE products	2025-11-10 14:36:50.405574
267	17	DELETE	admin/products	\N	admin DELETE products	2025-11-10 14:36:56.100021
268	17	DELETE	admin/products	\N	admin DELETE products	2025-11-10 14:37:13.110896
269	17	DELETE	product	8	Удален товар: ewds (ID: 8)	2025-11-10 14:46:32.001964
270	17	DELETE	admin/products	\N	admin DELETE products	2025-11-10 14:46:31.962651
271	17	UPDATE	product	1	Обновлен товар: Nike Air Max	2025-11-10 14:48:40.224682
272	17	UPDATE	admin/products	\N	admin UPDATE products	2025-11-10 14:48:40.18812
273	17	UPDATE	product	7	Обновлен товар: Adidas Samba	2025-11-10 14:49:33.373476
274	17	UPDATE	admin/products	\N	admin UPDATE products	2025-11-10 14:49:33.336674
275	17	UPDATE	product	2	Обновлен товар: Adidas Ultraboost	2025-11-10 14:50:41.12258
276	17	UPDATE	admin/products	\N	admin UPDATE products	2025-11-10 14:50:41.08902
277	17	UPDATE	product	3	Обновлен товар: Puma Runner	2025-11-10 14:51:33.776125
278	17	UPDATE	admin/products	\N	admin UPDATE products	2025-11-10 14:51:33.71986
279	17	UPDATE	product	4	Обновлен товар: Reebok Classic	2025-11-10 14:52:02.289488
280	17	UPDATE	admin/products	\N	admin UPDATE products	2025-11-10 14:52:02.225604
281	17	UPDATE	product	5	Обновлен товар: Nike Air Force 1	2025-11-10 14:52:10.374638
282	17	UPDATE	admin/products	\N	admin UPDATE products	2025-11-10 14:52:10.342533
283	17	UPDATE	product	3	Обновлен товар: Puma Runner	2025-11-10 14:52:15.682297
284	17	UPDATE	admin/products	\N	admin UPDATE products	2025-11-10 14:52:15.648403
285	17	UPDATE	product	2	Обновлен товар: Adidas Ultraboost	2025-11-10 14:52:20.168459
286	17	UPDATE	admin/products	\N	admin UPDATE products	2025-11-10 14:52:20.137244
287	24	CREATE	order	5	Создан заказ для пользователя ID: 24	2025-11-10 14:54:47.479594
288	24	CREATE	user/orders	\N	user CREATE orders	2025-11-10 14:54:45.145154
289	24	CREATE	user/order-products	\N	user CREATE order-products	2025-11-10 14:54:47.487422
290	24	REMOVE_FROM_BASKET	basket	14	Удален товар из корзины ID: 14	2025-11-10 14:54:47.513532
291	24	DELETE	user/14	\N	user DELETE 14	2025-11-10 14:54:47.512396
292	24	ADD_TO_BASKET	basket	15	Добавлен товар в корзину пользователя ID: 24	2025-11-10 16:27:45.373074
293	24	CREATE	user/basket	\N	user CREATE basket	2025-11-10 16:27:45.329557
294	24	CREATE	order	6	Создан заказ для пользователя ID: 24	2025-11-10 16:27:52.265447
296	24	CREATE	user/order-products	\N	user CREATE order-products	2025-11-10 16:27:52.30124
297	24	REMOVE_FROM_BASKET	basket	15	Удален товар из корзины ID: 15	2025-11-10 16:27:52.335648
298	24	DELETE	user/15	\N	user DELETE 15	2025-11-10 16:27:52.330911
\.


--
-- TOC entry 5050 (class 0 OID 25183)
-- Dependencies: 236
-- Data for Name: orderproducts; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.orderproducts (idorderproduct, idorder, idproductsize, quantity) FROM stdin;
4	1	1	1
5	1	2	1
6	2	5	1
7	3	70	2
8	4	50	1
9	5	49	1
10	6	10	1
\.


--
-- TOC entry 5048 (class 0 OID 25170)
-- Dependencies: 234
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.orders (idorder, iduser, orderdate) FROM stdin;
1	3	2025-09-28 16:27:41.528981
2	4	2025-09-28 16:27:41.528981
3	24	2025-10-28 14:32:20.181991
4	24	2025-10-30 12:49:39.169725
5	24	2025-11-10 14:54:45.145154
6	24	2025-11-10 16:27:49.593918
\.


--
-- TOC entry 5038 (class 0 OID 25084)
-- Dependencies: 224
-- Data for Name: products; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.products (idproduct, name, imageurl, price, idbrand, idcategory) FROM stdin;
1	Nike Air Max	https://images.laced.com/products/ada77c11-e055-475d-9934-c1bbc0acab2f.jpg	15000	1	1
7	Adidas Samba	https://spb-adidas.ru/image/cache/catalog/!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!1111111111111111111111111111111111111111111111111111/9777/adidas-originals-luzhniki-samba-classic-og-whitescarletclear-granite%20(1)-1000x600.jpg	10000	2	3
4	Reebok Classic	https://avatars.mds.yandex.net/i?id=1db813ddd4151330455c812c1eec6e6911a0f4f3-13108625-images-thumbs&n=13	13000	4	3
5	Nike Air Force 1	https://n.cdn.cdek.shopping/images/shopping/mUXWH8DnepkuarYX.jpg?v=1, https://media.s-bol.com/plNPw14GOQ8y/YE6lw9M/550x254.jpg, https://tse1.mm.bing.net/th/id/OIP.5akj-r0JjkkEzx1mVpsYSAHaEc?w=700&h=420&rs=1&pid=ImgDetMain&o=7&rm=3	35000	1	1
3	Puma Runner	https://avatars.mds.yandex.net/i?id=70ba1c7a33ef5d90499429c201277c0a53eb0d7e-5248224-images-thumbs&n=13	12000	3	3
2	Adidas Ultraboost	https://avatars.mds.yandex.net/i?id=d74183abb6a3625a74be13ec528cf4eca8a14501-5877187-images-thumbs&n=13	18000	2	1
\.


--
-- TOC entry 5040 (class 0 OID 25103)
-- Dependencies: 226
-- Data for Name: productsizes; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.productsizes (idproductsize, idproduct, size, quantity) FROM stdin;
61	7	36	100
62	7	37	100
63	7	38	100
64	7	39	100
3	1	38	100
4	1	39	100
6	1	41	100
7	1	42	100
8	1	43	100
9	1	44	100
11	2	36	100
12	2	37	100
13	2	38	100
14	2	39	100
15	2	40	100
16	2	41	100
17	2	42	100
18	2	43	100
19	2	44	100
20	2	45	100
21	3	36	100
22	3	37	100
23	3	38	100
24	3	39	100
25	3	40	100
26	3	41	100
27	3	42	100
28	3	43	100
29	3	44	100
30	3	45	100
31	4	36	100
32	4	37	100
33	4	38	100
34	4	39	100
35	4	40	100
36	4	41	100
37	4	42	100
38	4	43	100
39	4	44	100
40	4	45	100
41	5	36	100
42	5	37	100
43	5	38	100
44	5	39	100
45	5	40	100
46	5	41	100
47	5	42	100
48	5	43	100
1	1	36	100
2	1	37	100
5	1	40	100
65	7	40	100
66	7	41	100
67	7	42	100
68	7	43	100
69	7	44	100
70	7	45	98
50	5	45	99
49	5	44	99
10	1	45	99
\.


--
-- TOC entry 5052 (class 0 OID 25200)
-- Dependencies: 238
-- Data for Name: reports; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.reports (idreport, reportname, reporttype, reportdata, iduser, createdat) FROM stdin;
\.


--
-- TOC entry 5046 (class 0 OID 25149)
-- Dependencies: 232
-- Data for Name: reviews; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.reviews (idreview, idproduct, rating, comment, reviewdate, iduser) FROM stdin;
1	1	5	Отличные кроссовки!	2025-09-28 16:34:46.024037	3
2	2	4	Очень удобные	2025-09-28 16:34:46.024037	4
3	3	3	Нормально для пробежки	2025-09-28 16:34:46.024037	5
4	5	5	Классические Air Force 1, рекомендую	2025-09-28 16:34:46.024037	3
7	7	5	че за тяги	2025-10-28 15:04:21.617615	24
\.


--
-- TOC entry 5032 (class 0 OID 25058)
-- Dependencies: 218
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (iduser, fullname, email, passwordhash, roleid) FROM stdin;
1	Иван Иванов	admin@example.com	hashedpassword1	1
2	Мария Петрова	manager@example.com	hashedpassword2	2
3	Алексей Сидоров	user1@example.com	hashedpassword3	3
4	Ольга Смирнова	user2@example.com	hashedpassword4	3
5	Дмитрий Кузнецов	user3@example.com	hashedpassword5	3
10	Rusik	rus@rus.ru	passwordhash6	3
14	zxczxc	zxc@zxc.zxc	$2a$10$iKc0uuK9lOCjICWmIzT6iu14OEBqRu57RRC7jLDakQEIgHbQKXnW.	3
15	op	op@op.op	$2a$10$751kb2914A378xROAOcxKuE4dA5ewUvDS6W8761DNI16zoBKYgUhG	3
16	rus	rus@gmail.com	$2a$10$0Q6MYwO3PQ9m6dgfUPBQRODtMqS66EMBssd0ugdZhmRxEh8UZq6b2	3
17	admin	admin@gmail.com	$2a$10$IRhFP0CQtnDtNsx4um1Ov.o2CQm.ba.xLgLar58j6N30LysTf42MO	1
18	manager	manager@gmail.com	$2a$10$cgiVJg1b9q6fVNSZsXVbduAAd0Yy/XcBc.RVNCKd3lPaoLTbHTB.6	2
24	Musaev Rustam	rmusaevdg5@gmail.com	$2a$10$5CDwYAI13SFbRim7c5ioOO59afDqOkYNgMjSj5vRwupeCHtzN2cGK	3
\.


--
-- TOC entry 5072 (class 0 OID 0)
-- Dependencies: 227
-- Name: basket_idbasket_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.basket_idbasket_seq', 15, true);


--
-- TOC entry 5073 (class 0 OID 0)
-- Dependencies: 219
-- Name: brands_idbrand_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.brands_idbrand_seq', 6, true);


--
-- TOC entry 5074 (class 0 OID 0)
-- Dependencies: 221
-- Name: categories_idcategory_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.categories_idcategory_seq', 5, true);


--
-- TOC entry 5075 (class 0 OID 0)
-- Dependencies: 229
-- Name: favorites_idfavorites_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.favorites_idfavorites_seq', 9, true);


--
-- TOC entry 5076 (class 0 OID 0)
-- Dependencies: 239
-- Name: logs_idlog_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.logs_idlog_seq', 298, true);


--
-- TOC entry 5077 (class 0 OID 0)
-- Dependencies: 235
-- Name: orderproducts_idorderproduct_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.orderproducts_idorderproduct_seq', 10, true);


--
-- TOC entry 5078 (class 0 OID 0)
-- Dependencies: 233
-- Name: orders_idorder_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.orders_idorder_seq', 6, true);


--
-- TOC entry 5079 (class 0 OID 0)
-- Dependencies: 223
-- Name: products_idproduct_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.products_idproduct_seq', 8, true);


--
-- TOC entry 5080 (class 0 OID 0)
-- Dependencies: 225
-- Name: productsizes_idproductsize_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.productsizes_idproductsize_seq', 80, true);


--
-- TOC entry 5081 (class 0 OID 0)
-- Dependencies: 237
-- Name: reports_idreport_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.reports_idreport_seq', 1, false);


--
-- TOC entry 5082 (class 0 OID 0)
-- Dependencies: 231
-- Name: reviews_idreview_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.reviews_idreview_seq', 7, true);


--
-- TOC entry 5083 (class 0 OID 0)
-- Dependencies: 217
-- Name: users_iduser_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.users_iduser_seq', 24, true);


--
-- TOC entry 4852 (class 2606 OID 25120)
-- Name: basket basket_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.basket
    ADD CONSTRAINT basket_pkey PRIMARY KEY (idbasket);


--
-- TOC entry 4844 (class 2606 OID 25074)
-- Name: brands brands_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.brands
    ADD CONSTRAINT brands_pkey PRIMARY KEY (idbrand);


--
-- TOC entry 4846 (class 2606 OID 25082)
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (idcategory);


--
-- TOC entry 4854 (class 2606 OID 25137)
-- Name: favorites favorites_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_pkey PRIMARY KEY (idfavorites);


--
-- TOC entry 4864 (class 2606 OID 25223)
-- Name: logs logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_pkey PRIMARY KEY (idlog);


--
-- TOC entry 4860 (class 2606 OID 25188)
-- Name: orderproducts orderproducts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orderproducts
    ADD CONSTRAINT orderproducts_pkey PRIMARY KEY (idorderproduct);


--
-- TOC entry 4858 (class 2606 OID 25176)
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (idorder);


--
-- TOC entry 4848 (class 2606 OID 25091)
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (idproduct);


--
-- TOC entry 4850 (class 2606 OID 25108)
-- Name: productsizes productsizes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.productsizes
    ADD CONSTRAINT productsizes_pkey PRIMARY KEY (idproductsize);


--
-- TOC entry 4862 (class 2606 OID 25208)
-- Name: reports reports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_pkey PRIMARY KEY (idreport);


--
-- TOC entry 4856 (class 2606 OID 25158)
-- Name: reviews reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_pkey PRIMARY KEY (idreview);


--
-- TOC entry 4840 (class 2606 OID 25067)
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- TOC entry 4842 (class 2606 OID 25065)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (iduser);


--
-- TOC entry 4879 (class 2620 OID 25268)
-- Name: products trg_add_default_sizes; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_add_default_sizes AFTER INSERT ON public.products FOR EACH ROW EXECUTE FUNCTION public.adddefaultsizes();


--
-- TOC entry 4880 (class 2620 OID 33250)
-- Name: productsizes trg_check_product_quantity; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_check_product_quantity BEFORE INSERT OR UPDATE ON public.productsizes FOR EACH ROW EXECUTE FUNCTION public.check_product_quantity();


--
-- TOC entry 4881 (class 2620 OID 25252)
-- Name: orderproducts trg_decrease_stock; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_decrease_stock AFTER INSERT ON public.orderproducts FOR EACH ROW EXECUTE FUNCTION public.decreasestock();


--
-- TOC entry 4868 (class 2606 OID 25126)
-- Name: basket basket_idproductsize_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.basket
    ADD CONSTRAINT basket_idproductsize_fkey FOREIGN KEY (idproductsize) REFERENCES public.productsizes(idproductsize);


--
-- TOC entry 4869 (class 2606 OID 25121)
-- Name: basket basket_iduser_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.basket
    ADD CONSTRAINT basket_iduser_fkey FOREIGN KEY (iduser) REFERENCES public.users(iduser);


--
-- TOC entry 4870 (class 2606 OID 25143)
-- Name: favorites favorites_idproductsize_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_idproductsize_fkey FOREIGN KEY (idproductsize) REFERENCES public.productsizes(idproductsize);


--
-- TOC entry 4871 (class 2606 OID 25138)
-- Name: favorites favorites_iduser_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_iduser_fkey FOREIGN KEY (iduser) REFERENCES public.users(iduser);


--
-- TOC entry 4878 (class 2606 OID 25224)
-- Name: logs logs_iduser_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_iduser_fkey FOREIGN KEY (iduser) REFERENCES public.users(iduser);


--
-- TOC entry 4875 (class 2606 OID 25189)
-- Name: orderproducts orderproducts_idorder_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orderproducts
    ADD CONSTRAINT orderproducts_idorder_fkey FOREIGN KEY (idorder) REFERENCES public.orders(idorder);


--
-- TOC entry 4876 (class 2606 OID 25194)
-- Name: orderproducts orderproducts_idproductsize_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orderproducts
    ADD CONSTRAINT orderproducts_idproductsize_fkey FOREIGN KEY (idproductsize) REFERENCES public.productsizes(idproductsize);


--
-- TOC entry 4874 (class 2606 OID 25177)
-- Name: orders orders_iduser_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_iduser_fkey FOREIGN KEY (iduser) REFERENCES public.users(iduser);


--
-- TOC entry 4865 (class 2606 OID 25092)
-- Name: products products_idbrand_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_idbrand_fkey FOREIGN KEY (idbrand) REFERENCES public.brands(idbrand);


--
-- TOC entry 4866 (class 2606 OID 25097)
-- Name: products products_idcategory_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_idcategory_fkey FOREIGN KEY (idcategory) REFERENCES public.categories(idcategory);


--
-- TOC entry 4867 (class 2606 OID 25109)
-- Name: productsizes productsizes_idproduct_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.productsizes
    ADD CONSTRAINT productsizes_idproduct_fkey FOREIGN KEY (idproduct) REFERENCES public.products(idproduct);


--
-- TOC entry 4877 (class 2606 OID 25209)
-- Name: reports reports_iduser_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_iduser_fkey FOREIGN KEY (iduser) REFERENCES public.users(iduser);


--
-- TOC entry 4872 (class 2606 OID 25159)
-- Name: reviews reviews_idproduct_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_idproduct_fkey FOREIGN KEY (idproduct) REFERENCES public.products(idproduct);


--
-- TOC entry 4873 (class 2606 OID 25164)
-- Name: reviews reviews_iduser_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_iduser_fkey FOREIGN KEY (iduser) REFERENCES public.users(iduser);


-- Completed on 2025-11-10 16:43:46

--
-- PostgreSQL database dump complete
--

